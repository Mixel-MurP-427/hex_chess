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

// These types are dumb lol. I'll probably replace them, idk.

type jsonCoordinate struct {
	X uint8 `json:"x"`
	Y uint8 `json:"y"`
}

type pieceJSON struct {
	Color  bool  `json:"color"`
	Angles uint8 `json:"angles"`
}

type squareJSON struct {
	Position jsonCoordinate `json:"position"`
	Piece    pieceJSON      `json:"piece"`
}

type sourceMovesJSON struct {
	From jsonCoordinate   `json:"from"`
	To   []jsonCoordinate `json:"to"`
}

type kingPositionJSON [3]jsonCoordinate

type stateJSON struct {
	Type           string             `json:"type"`
	Revision       uint64             `json:"revision"`
	Board          []squareJSON       `json:"board"`
	Turn           bool               `json:"turn"`
	TrueKing       kingPositionJSON   `json:"trueKing"`
	FalseKing      kingPositionJSON   `json:"falseKing"`
	LegalMoves     []sourceMovesJSON  `json:"legalMoves"`
	LegalKingMoves []kingPositionJSON `json:"legalKingMoves"`
}

type errorJSON struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type clientMessage struct {
	Type     string           `json:"type"`
	From     jsonCoordinate   `json:"from"`
	To       jsonCoordinate   `json:"to"`
	Rotation uint8            `json:"rotation"`
	King     []jsonCoordinate `json:"king"`
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
	trueKing  KingNeighbors
	falseKing KingNeighbors
	turn      bool
	revision  uint64
	clients   map[*wsClient]struct{}
}

func newGameServer() *gameServer {
	return &gameServer{
		board:     SetupBoard(false),
		trueKing:  TheTrueKing.getKingNeighbors(),
		falseKing: TheFalseKing.getKingNeighbors(),
		turn:      true, // True player goes first
		clients:   make(map[*wsClient]struct{}),
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

func coordJSONFromCoord(coord Coord) jsonCoordinate {
	return jsonCoordinate{X: coord.X, Y: coord.Y}
}

func kingPositionJSONFromKing(position KingNeighbors) kingPositionJSON {
	return kingPositionJSON{
		coordJSONFromCoord(position[0]),
		coordJSONFromCoord(position[1]),
		coordJSONFromCoord(position[2]),
	}
}

func coordFromJSON(coord jsonCoordinate) Coord {
	return Coord{coord.X, coord.Y}
}

func kingPositionFromJSON(position []jsonCoordinate) (KingNeighbors, error) {
	if len(position) != 3 {
		return KingNeighbors{}, errors.New("king position must contain exactly three vertices")
	}
	return KingNeighbors{
		coordFromJSON(position[0]),
		coordFromJSON(position[1]),
		coordFromJSON(position[2]),
	}, nil
}

func (s *gameServer) stateLocked() (stateJSON, error) {
	state := stateJSON{
		Type:           "state",
		Revision:       s.revision,
		Board:          make([]squareJSON, 0, 61),
		Turn:           s.turn,
		TrueKing:       kingPositionJSONFromKing(s.trueKing),
		FalseKing:      kingPositionJSONFromKing(s.falseKing),
		LegalMoves:     make([]sourceMovesJSON, 0),
		LegalKingMoves: make([]kingPositionJSON, 0),
	}
	for x, row := range s.board {
		ys := make([]int, 0, len(row))
		for y := range row {
			ys = append(ys, int(y))
		}
		slices.Sort(ys)
		for _, rawY := range ys {
			piece := row[uint8(rawY)]
			angles := piece.Ang

			state.Board = append(state.Board, squareJSON{
				Position: jsonCoordinate{X: uint8(x), Y: uint8(rawY)},
				Piece: pieceJSON{
					Color:  piece.Col,
					Angles: angles,
				},
			})
			if piece.Ang == 0 || piece.Col != s.turn {
				continue
			}
			moves, err := s.board.GetLegalMoves(Coord{uint8(x), uint8(rawY)})
			if err != nil {
				return stateJSON{}, fmt.Errorf("calculate legal moves at (%d, %d): %w", x, rawY, err)
			}
			destinations := make([]jsonCoordinate, 0, len(moves))
			king := s.trueKing
			//opponentKing := s.falseKing
			if !s.turn {
				king = s.falseKing
				//opponentKing = s.trueKing
			}
			for _, move := range moves {
				nextBoard := s.board.clone()
				if err := nextBoard.Move(Coord{uint8(x), uint8(rawY)}, move, 0); err != nil { // I hotwired that zero in there. It will probably cause bugs later.
					return stateJSON{}, fmt.Errorf("validate move from (%d, %d) to (%d, %d): %w", x, rawY, move.X, move.Y, err)
				}
				inCheck, err := nextBoard.Check(king, s.turn)
				if err != nil {
					return stateJSON{}, fmt.Errorf("check king safety after move: %w", err)
				}
				if inCheck {
					continue
				}
				destinations = append(destinations, coordJSONFromCoord(move))
			}
			state.LegalMoves = append(state.LegalMoves, sourceMovesJSON{
				From: jsonCoordinate{X: uint8(x), Y: uint8(rawY)},
				To:   destinations,
			})
		}
	}
	activeKing := s.trueKing
	opponentKing := s.falseKing
	if !s.turn {
		activeKing = s.falseKing
		opponentKing = s.trueKing
	}
	kingMoves, err := s.board.KingMovesAgainst(activeKing, s.turn, opponentKing)
	if err != nil {
		return stateJSON{}, fmt.Errorf("calculate legal king moves: %w", err)
	}
	for _, move := range kingMoves {
		state.LegalKingMoves = append(state.LegalKingMoves, kingPositionJSONFromKing(move))
	}
	return state, nil
}

func (s *gameServer) process(message clientMessage) error {
	s.mu.Lock()
	oldBoard, oldTrueKing, oldFalseKing := s.board, s.trueKing, s.falseKing
	oldTurn, oldRevision := s.turn, s.revision
	var err error
	switch message.Type {
	case "move":
		from, to := coordFromJSON(message.From), coordFromJSON(message.To)
		if !posExists(from) || !posExists(to) {
			err = errors.New("move coordinates must be vertices on the board")
			break
		}
		piece := s.board[from.X][from.Y]
		if piece.Ang == 0 {
			err = errors.New("there is no piece at the selected vertex")
			break
		}
		if piece.Col != s.turn {
			err = errors.New("you can only move a piece belonging to the active side")
			break
		}
		nextBoard := s.board.clone()
		if err = nextBoard.Move(from, to, message.Rotation); err != nil {
			break
		}
		king := s.trueKing
		//opponentKing := s.falseKing
		if !s.turn {
			king = s.falseKing
			//opponentKing = s.trueKing
		}
		var inCheck bool
		inCheck, err = nextBoard.Check(king, s.turn)
		if err != nil {
			break
		}
		if inCheck {
			err = errors.New("move would leave your king in check")
			break
		}
		s.board = nextBoard
		s.turn = !s.turn
		s.revision++
	case "king_move":
		nextKing, parseErr := kingPositionFromJSON(message.King)
		if parseErr != nil {
			err = parseErr
			break
		}
		currentKing := s.trueKing
		opponentKing := s.falseKing
		if !s.turn {
			currentKing = s.falseKing
			opponentKing = s.trueKing
		}
		if err = s.board.MoveKingAgainst(currentKing, nextKing, s.turn, opponentKing); err != nil {
			break
		}
		if s.turn {
			s.trueKing = canonicalKingPosition(nextKing)
		} else {
			s.falseKing = canonicalKingPosition(nextKing)
		}
		s.turn = !s.turn
		s.revision++
	default:
		err = fmt.Errorf("unknown message type %q", message.Type)
	}
	if err != nil {
		s.mu.Unlock()
		return err
	}
	state, err := s.stateLocked()
	if err != nil {
		s.board, s.trueKing, s.falseKing = oldBoard, oldTrueKing, oldFalseKing
		s.turn, s.revision = oldTurn, oldRevision
		s.mu.Unlock()
		return err
	}
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
	conn, err := websocket.Upgrade(writer, request, nil, 1024, 1024)
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
