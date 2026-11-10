package hexchess

import (
	"slices"
	"testing"
)

func TestMoveWithRotationAcrossDirectionBoundary(t *testing.T) {
	board := SetupBoard(true)
	board[2][2] = Piece{col: true, ang: []uint8{1, 4}}

	if err := board.MoveWithRotation(Coord{2, 2}, Coord{3, 2}, 2); err != nil {
		t.Fatalf("MoveWithRotation returned error: %v", err)
	}
	if got := board[3][2].ang; !slices.Equal(got, []uint8{0, 3}) {
		t.Fatalf("rotated angles = %v, want [0 3]", got)
	}
	if len(board[2][2].ang) != 0 {
		t.Fatalf("source vertex still contains a piece: %v", board[2][2])
	}
}

func TestMoveWithRotationRejectsInvalidMoveWithoutMutation(t *testing.T) {
	board := SetupBoard(true)
	board[2][2] = Piece{col: true, ang: []uint8{1}}

	if err := board.MoveWithRotation(Coord{2, 2}, Coord{2, 3}, 0); err == nil {
		t.Fatal("MoveWithRotation accepted a destination outside the piece's direction")
	}
	if len(board[2][2].ang) == 0 || len(board[2][3].ang) != 0 {
		t.Fatal("invalid move mutated the board")
	}
}

func TestKingMovesToAdjacentSafeTriangles(t *testing.T) {
	board := SetupBoard(true)
	moves, err := board.KingMoves(tKStart, true)
	if err != nil {
		t.Fatalf("KingMoves returned error: %v", err)
	}
	if len(moves) != 3 {
		t.Fatalf("KingMoves returned %d destinations, want 3: %v", len(moves), moves)
	}
	if err := board.MoveKing(tKStart, moves[0], true); err != nil {
		t.Fatalf("MoveKing rejected a legal destination: %v", err)
	}
}

func TestKingMovesRejectInvalidTriangle(t *testing.T) {
	board := SetupBoard(true)
	invalid := KingPos{Coord{2, 2}, Coord{4, 2}, Coord{2, 3}}
	if _, err := board.KingMoves(invalid, true); err == nil {
		t.Fatal("KingMoves accepted a non-adjacent king triangle")
	}
}

func TestKingMoveAllowsOccupiedVertices(t *testing.T) {
	board := SetupBoard(true)
	moves, err := board.KingMoves(tKStart, true)
	if err != nil {
		t.Fatalf("KingMoves returned error: %v", err)
	}
	move := moves[0]
	for _, vertex := range move {
		if !slices.Contains(tKStart[:], vertex) {
			board[vertex.x][vertex.y] = Piece{col: true, ang: []uint8{1}}
			break
		}
	}
	if err := board.MoveKing(tKStart, move, true); err != nil {
		t.Fatalf("MoveKing rejected a safe triangle with an occupied vertex: %v", err)
	}
}

func TestKingMovesExcludeCheckedDestinations(t *testing.T) {
	board := SetupBoard(true)
	adjacent, err := board.adjacentKingPositions(tKStart)
	if err != nil {
		t.Fatalf("adjacentKingPositions returned error: %v", err)
	}
	foundThreat := false
	for x, row := range board {
		for y := range row {
			candidate := Coord{uint8(x), y}
			if slices.Contains(tKStart[:], candidate) {
				continue
			}
			for angle := uint8(0); angle < 6; angle++ {
				board[x][y] = Piece{col: false, ang: []uint8{angle}}
				sourceInCheck, checkErr := board.Check(tKStart, true)
				if checkErr != nil {
					t.Fatalf("Check on source king: %v", checkErr)
				}
				if sourceInCheck {
					board[x][y] = Piece{}
					continue
				}
				for _, destination := range adjacent {
					inCheck, checkErr := board.Check(destination, true)
					if checkErr != nil {
						t.Fatalf("Check on destination king: %v", checkErr)
					}
					if !inCheck {
						continue
					}
					legal, moveErr := board.KingMoves(tKStart, true)
					if moveErr != nil {
						t.Fatalf("KingMoves returned error: %v", moveErr)
					}
					if slices.Contains(legal, destination) {
						t.Fatalf("KingMoves included checked destination %v", destination)
					}
					foundThreat = true
					break
				}
				board[x][y] = Piece{}
				if foundThreat {
					break
				}
			}
			if foundThreat {
				break
			}
		}
		if foundThreat {
			break
		}
	}
	if !foundThreat {
		t.Fatal("could not construct a king destination threatened by one enemy piece")
	}
}
