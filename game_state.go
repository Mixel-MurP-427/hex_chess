package hexchess

import (
	"fmt"
	"slices"
)

func cloneBoard(board Board) Board {
	var cloned Board
	for x, row := range board {
		cloned[x] = make(map[uint8]Piece, len(row))
		for y, piece := range row {
			cloned[x][y] = Piece{col: piece.col, ang: slices.Clone(piece.ang)}
		}
	}
	return cloned
}

func (b *Board) MoveWithRotation(from, to Coord, rotation uint8) error {
	if rotation > 5 {
		return fmt.Errorf("invalid rotation %d: must be between 0 and 5", rotation)
	}
	piece, ok := b[from.x][from.y]
	if !ok || len(piece.ang) == 0 {
		return fmt.Errorf("no piece at (%d, %d)", from.x, from.y)
	}
	return b.Move(from, to, rotatePiece(piece, rotation))
}

var kingDirections = [6][2]int{
	{1, -1}, {1, 0}, {0, 1}, {-1, 1}, {-1, 0}, {0, -1},
}

func (b *Board) neighbor(c Coord, direction uint8) (Coord, bool) {
	delta := kingDirections[direction]
	x, y := int(c.x)+delta[0], int(c.y)+delta[1]
	if x < 0 || x > 8 || y < 0 || y > 8 {
		return Coord{}, false
	}
	next := Coord{uint8(x), uint8(y)}
	return next, b.posExists(next)
}

func (b *Board) areNeighbors(a, c Coord) bool {
	for direction := range kingDirections {
		next, ok := b.neighbor(a, uint8(direction))
		if ok && next == c {
			return true
		}
	}
	return false
}

func (b *Board) validKingPosition(position KingPos) bool {
	for i, vertex := range position {
		if !b.posExists(vertex) {
			return false
		}
		for j := 0; j < i; j++ {
			if vertex == position[j] || !b.areNeighbors(vertex, position[j]) {
				return false
			}
		}
	}
	return true
}

func canonicalKingPosition(position KingPos) KingPos {
	slices.SortFunc(position[:], func(a, b Coord) int {
		if a.x != b.x {
			return int(a.x) - int(b.x)
		}
		return int(a.y) - int(b.y)
	})
	return position
}

func (b *Board) adjacentKingPositions(position KingPos) ([]KingPos, error) {
	if !b.validKingPosition(position) {
		return nil, fmt.Errorf("invalid king position: %v", position)
	}
	unique := make(map[KingPos]struct{})
	for i := 0; i < len(position); i++ {
		for j := i + 1; j < len(position); j++ {
			first, second := position[i], position[j]
			for direction := range kingDirections {
				third, ok := b.neighbor(first, uint8(direction))
				if !ok || third == second || third == position[0] || third == position[1] || third == position[2] {
					continue
				}
				if !b.areNeighbors(second, third) {
					continue
				}
				candidate := canonicalKingPosition(KingPos{first, second, third})
				unique[candidate] = struct{}{}
			}
		}
	}
	moves := make([]KingPos, 0, len(unique))
	for move := range unique {
		moves = append(moves, move)
	}
	slices.SortFunc(moves, func(a, c KingPos) int {
		for i := range a {
			if a[i].x != c[i].x {
				return int(a[i].x) - int(c[i].x)
			}
			if a[i].y != c[i].y {
				return int(a[i].y) - int(c[i].y)
			}
		}
		return 0
	})
	return moves, nil
}

func (b *Board) KingMoves(position KingPos, color bool) ([]KingPos, error) {
	adjacent, err := b.adjacentKingPositions(position)
	if err != nil {
		return nil, err
	}
	moves := make([]KingPos, 0, len(adjacent))
	for _, move := range adjacent {
		inCheck, checkErr := b.Check(move, color)
		if checkErr != nil {
			return nil, checkErr
		}
		if !inCheck {
			moves = append(moves, move)
		}
	}
	return moves, nil
}

func (b *Board) kingInCheck(position KingPos, color bool, opponent KingPos) (bool, error) {
	inCheck, err := b.Check(position, color)
	if err != nil || inCheck {
		return inCheck, err
	}
	opponentMoves, err := b.adjacentKingPositions(opponent)
	if err != nil {
		return false, err
	}
	position = canonicalKingPosition(position)
	return position == canonicalKingPosition(opponent) || slices.Contains(opponentMoves, position), nil
}

func (b *Board) KingMovesAgainst(position KingPos, color bool, opponent KingPos) ([]KingPos, error) {
	moves, err := b.KingMoves(position, color)
	if err != nil {
		return nil, err
	}
	opponentMoves, err := b.adjacentKingPositions(opponent)
	if err != nil {
		return nil, err
	}
	opponent = canonicalKingPosition(opponent)
	safeMoves := make([]KingPos, 0, len(moves))
	for _, move := range moves {
		if move != opponent && !slices.Contains(opponentMoves, move) {
			safeMoves = append(safeMoves, move)
		}
	}
	return safeMoves, nil
}

func (b *Board) MoveKing(current, next KingPos, color bool) error {
	moves, err := b.KingMoves(current, color)
	if err != nil {
		return err
	}
	if !slices.Contains(moves, canonicalKingPosition(next)) {
		return fmt.Errorf("king cannot move from %v to %v", current, next)
	}
	return nil
}

func (b *Board) MoveKingAgainst(current, next KingPos, color bool, opponent KingPos) error {
	moves, err := b.KingMovesAgainst(current, color, opponent)
	if err != nil {
		return err
	}
	if !slices.Contains(moves, canonicalKingPosition(next)) {
		return fmt.Errorf("king cannot move from %v to %v", current, next)
	}
	return nil
}
