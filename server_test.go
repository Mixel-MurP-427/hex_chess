package hexchess

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestServerProcessMoveAndEnforceTurn(t *testing.T) {
	server := newGameServer()
	server.board = SetupBoard(true)
	server.board[0][4] = Piece{col: true, ang: []uint8{1}}

	if err := server.process(clientMessage{
		Type: "move", From: coordinateJSON{X: 0, Y: 4}, To: coordinateJSON{X: 1, Y: 4},
	}); err != nil {
		t.Fatalf("process valid move: %v", err)
	}
	if server.turn || server.revision != 1 {
		t.Fatalf("turn = %t, revision = %d; want false, 1", server.turn, server.revision)
	}

	if err := server.process(clientMessage{
		Type: "move", From: coordinateJSON{X: 0, Y: 4}, To: coordinateJSON{X: 1, Y: 4},
	}); err == nil {
		t.Fatal("process accepted a second move from the previous side")
	}
	if server.turn || server.revision != 1 {
		t.Fatalf("rejected move changed state: turn = %t, revision = %d", server.turn, server.revision)
	}
}

func TestServerRejectsMovingWrongColor(t *testing.T) {
	server := newGameServer()
	err := server.process(clientMessage{
		Type: "move", From: coordinateJSON{X: 7, Y: 5}, To: coordinateJSON{X: 6, Y: 5},
	})
	if err == nil {
		t.Fatal("server accepted a move for the inactive side")
	}
	if server.revision != 0 || !server.turn {
		t.Fatalf("rejected move changed game state: turn = %t, revision = %d", server.turn, server.revision)
	}
}

func TestServerProcessesKingMove(t *testing.T) {
	server := newGameServer()
	server.board = SetupBoard(true)
	moves, err := server.board.KingMovesAgainst(server.trueKing, true, server.falseKing)
	if err != nil {
		t.Fatalf("calculate king moves: %v", err)
	}
	if len(moves) == 0 {
		t.Fatal("expected at least one legal king move")
	}
	target := moves[0]
	if err := server.process(clientMessage{
		Type: "king_move",
		King: []coordinateJSON{
			coordJSONFromCoord(target[0]),
			coordJSONFromCoord(target[1]),
			coordJSONFromCoord(target[2]),
		},
	}); err != nil {
		t.Fatalf("process king move: %v", err)
	}
	if server.turn || server.revision != 1 || server.trueKing != target {
		t.Fatalf("king move did not update state: turn=%t revision=%d king=%v", server.turn, server.revision, server.trueKing)
	}
}

func TestNewHandlerServesUIAndWebSocketState(t *testing.T) {
	httpServer := httptest.NewServer(NewHandler())
	defer httpServer.Close()

	response, err := httpServer.Client().Get(httpServer.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatalf("GET / returned status %d", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read UI response: %v", err)
	}
	if !strings.Contains(string(body), "Hex Chess") {
		t.Fatal("UI response does not contain the game title")
	}

	websocketURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		t.Fatalf("connect WebSocket: %v", err)
	}
	defer conn.Close()
	var state stateJSON
	if err := conn.ReadJSON(&state); err != nil {
		t.Fatalf("read initial state: %v", err)
	}
	if state.Type != "state" || len(state.Board) != 61 || !state.Turn {
		t.Fatalf("unexpected initial state: type=%q vertices=%d turn=%t", state.Type, len(state.Board), state.Turn)
	}
	foundEmptyVertex := false
	for _, square := range state.Board {
		if len(square.Piece.Angles) == 0 {
			foundEmptyVertex = true
			if square.Piece.Angles == nil {
				t.Fatal("empty piece angles must be an empty array, not null")
			}
			break
		}
	}
	if !foundEmptyVertex {
		t.Fatal("initial board contains no empty vertices")
	}
	if state.LegalMoves == nil || state.LegalKingMoves == nil {
		t.Fatal("legal move collections must be arrays, not null")
	}
	if err := conn.WriteJSON(clientMessage{Type: "unknown"}); err != nil {
		t.Fatalf("send invalid message: %v", err)
	}
	var responseMessage errorJSON
	if err := conn.ReadJSON(&responseMessage); err != nil {
		t.Fatalf("read invalid-message response: %v", err)
	}
	if responseMessage.Type != "error" || responseMessage.Message == "" {
		t.Fatalf("unexpected invalid-message response: %+v", responseMessage)
	}
}
