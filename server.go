package hexchess

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
)

//go:embed web
var webAssets embed.FS

type squareJSON struct {
	Position Coord `json:"position"`
	Piece    Piece `json:"piece"`
}

type sourceMovesJSON map[string][]Coord

type stateJSON struct {
	Type           string          `json:"type"`
	Revision       uint64          `json:"revision"`
	Board          []squareJSON    `json:"board"`
	Turn           bool            `json:"turn"`
	TrueKing       KingPos         `json:"trueKing"`
	FalseKing      KingPos         `json:"falseKing"`
	LegalMoves     sourceMovesJSON `json:"legalMoves"`
	LegalKingMoves []KingPos       `json:"legalKingMoves"`
}

type errorJSON struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type clientMessage struct {
	Type     string  `json:"type"`
	From     Coord   `json:"from"`
	To       Coord   `json:"to"`
	Rotation uint8   `json:"rotation"`
	King     KingPos `json:"king"`
}

type wsClient struct {
	conn    *websocket.Conn
	writeMu sync.Mutex
}

func (c *wsClient) writeJSON(value any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteJSON(value)
}

type gameServer struct {
	mu        sync.Mutex
	board     Board
	trueKing  KingPos
	falseKing KingPos
	turn      bool
	revision  uint64 // turn number
	clients   map[*wsClient]struct{}
}

func newGameServer() *gameServer {
	return &gameServer{
		board:     SetupBoard(false),
		trueKing:  InitKing(true),
		falseKing: InitKing(false),
		turn:      true, // True player goes first
		clients:   make(map[*wsClient]struct{}),
	}
}

func (s *gameServer) getActiveKing() KingPos {
	if s.turn {
		return s.trueKing
	} else {
		return s.falseKing
	}
}
func (s *gameServer) getPointerToActiveKing() *KingPos {
	if s.turn {
		return &s.trueKing
	} else {
		return &s.falseKing
	}
}

// NewHandler creates the local game server and serves its browser UI.
func NewHandler() http.Handler {
	server := newGameServer()
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", server.serveWebSocket)
	web, err := fs.Sub(webAssets, "web")
	if err != nil {
		panic(fmt.Errorf("load embedded web assets: %w", err))
	}
	mux.Handle("/", http.FileServer(http.FS(web)))
	return mux
}

// Provides available actions to the current player
func (s *gameServer) stateLocked() (stateJSON, error) {
	state := stateJSON{
		Type:           "state",
		Revision:       s.revision,
		Board:          make([]squareJSON, 0, 61),
		Turn:           s.turn,
		TrueKing:       s.trueKing,
		FalseKing:      s.falseKing,
		LegalMoves:     make(sourceMovesJSON, 11),
		LegalKingMoves: make([]KingPos, 0),
	}

	activeKing := s.getActiveKing()

	// Get available moves that each piece can make

	for xInt, row := range s.board {
		x := uint8(xInt)
		// ys is the list of y values in each x row
		ys := make([]uint8, 0, len(row))
		for y := range row {
			ys = append(ys, y)
		}
		// Sorting is unnecessary because they already come sorted?
		slices.Sort(ys)
		// Iterate over all spaces on the board
		for _, rawY := range ys {
			piece := row[rawY]

			// Add this space to the board state
			state.Board = append(state.Board, squareJSON{
				Position: Coord{x, rawY},
				Piece:    Piece{piece.Col, piece.Ang},
			})
			// If the player is not allowed to manipulate this piece
			if piece.Ang == 0 || piece.Col != s.turn {
				continue
			}
			// Get legal moves for this piece
			moves, err := s.board.GetLegalMoves(Coord{x, rawY}, true, activeKing)
			if err != nil {
				return stateJSON{}, fmt.Errorf("calculate legal moves at (%d, %d): %w", x, rawY, err)
			}
			state.LegalMoves[fmt.Sprintf("%d,%d", x, rawY)] = moves
		}
	}

	// Get king moves

	cMoves, err := activeKing.GetLegalMoves(&s.board, true)
	if err != nil {
		return stateJSON{}, fmt.Errorf("calculate legal king moves: %w", err)
	}

	state.LegalKingMoves = []KingPos{}
	for _, move := range cMoves {
		state.LegalKingMoves = append(state.LegalKingMoves, KingPos{move, !activeKing.Ontop, activeKing.Col})
	}
	return state, nil
}

// Process move sent by the player
func (s *gameServer) process(message clientMessage) error {
	s.mu.Lock()
	oldBoard, oldTrueKing, oldFalseKing := s.board, s.trueKing, s.falseKing
	oldTurn, oldRevision := s.turn, s.revision
	var err error
	activeKing := s.getPointerToActiveKing() // TODO test this
	// read message
	switch message.Type {
	// If a normal piece was moved
	case "move":
		from, to := message.From, message.To

		// Validation
		if s.board[from.X][from.Y].Col != s.turn {
			err = errors.New("you can only move a piece belonging to the active side")
			break
		}

		// Advance the game
		err = s.board.Move(from, to, message.Rotation, *activeKing)
		if err != nil {
			break
		}
		s.turn = !s.turn
		s.revision++

	// if a king was moved
	case "king_move":

		// Try moving the king
		if err = activeKing.Move(&s.board, message.King.KCoord); err != nil {
			break
		}

		// Advance the game
		s.turn = !s.turn
		s.revision++

	default:
		err = fmt.Errorf("unknown message type %q", message.Type)
	} // End of switch block

	// Return any errors caught
	if err != nil {
		s.mu.Unlock()
		return err
	}

	// Set up next turn
	state, err := s.stateLocked()
	if err != nil {
		// Reset state if error
		s.board, s.trueKing, s.falseKing = oldBoard, oldTrueKing, oldFalseKing
		s.turn, s.revision = oldTurn, oldRevision
		s.mu.Unlock()
		return err
	}

	// Do clienty stuff
	clients := make([]*wsClient, 0, len(s.clients))
	for client := range s.clients {
		clients = append(clients, client)
	}
	s.mu.Unlock()

	for _, client := range clients {
		if writeErr := client.writeJSON(state); writeErr != nil {
			log.Printf("broadcast game state: %v", writeErr)
			s.removeClient(client)
		}
	}
	return nil
}

func (s *gameServer) removeClient(client *wsClient) {
	s.mu.Lock()
	_, exists := s.clients[client]
	delete(s.clients, client)
	s.mu.Unlock()
	if exists {
		if err := client.conn.Close(); err != nil {
			log.Printf("close WebSocket connection: %v", err)
		}
	}
}

func (s *gameServer) serveWebSocket(writer http.ResponseWriter, request *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Printf("upgrade WebSocket connection: %v", err)
		return
	}
	client := &wsClient{conn: conn}
	conn.SetReadLimit(64 * 1024)
	s.mu.Lock()
	s.clients[client] = struct{}{}
	state, err := s.stateLocked()
	s.mu.Unlock()
	if err != nil {
		log.Printf("create initial game state: %v", err)
		s.removeClient(client)
		return
	}
	if err := client.writeJSON(state); err != nil {
		log.Printf("send initial game state: %v", err)
		s.removeClient(client)
		return
	}

	for {
		var message clientMessage
		if err := conn.ReadJSON(&message); err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("read WebSocket message: %v", err)
			}
			break
		}
		if err := s.process(message); err != nil {
			if writeErr := client.writeJSON(errorJSON{Type: "error", Message: err.Error()}); writeErr != nil {
				log.Printf("send WebSocket error response: %v", writeErr)
				break
			}
		}
	}
	s.removeClient(client)
}
