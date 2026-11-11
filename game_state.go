package hexchess

// import (
// 	"fmt"
// 	"slices"
// )

// var kingDirections = [6][2]int{
// 	{1, -1}, {1, 0}, {0, 1}, {-1, 1}, {-1, 0}, {0, -1},
// }

// func (b *Board) neighbor(c Coord, direction uint8) (Coord, bool) {
// 	delta := kingDirections[direction]
// 	x, y := int(c.x)+delta[0], int(c.y)+delta[1]
// 	if x < 0 || x > 8 || y < 0 || y > 8 {
// 		return Coord{}, false
// 	}
// 	next := Coord{uint8(x), uint8(y)}
// 	return next, b.posExists(next)
// }

// func (b *Board) areNeighbors(a, c Coord) bool {
// 	for direction := range kingDirections {
// 		next, ok := b.neighbor(a, uint8(direction))
// 		if ok && next == c {
// 			return true
// 		}
// 	}
// 	return false
// }

// func (b *Board) validKingPosition(position KingNeighbors) bool {
// 	for i, vertex := range position {
// 		if !b.posExists(vertex) {
// 			return false
// 		}
// 		for j := 0; j < i; j++ {
// 			if vertex == position[j] || !b.areNeighbors(vertex, position[j]) {
// 				return false
// 			}
// 		}
// 	}
// 	return true
// }

// func canonicalKingPosition(position KingNeighbors) KingNeighbors {
// 	slices.SortFunc(position[:], func(a, b Coord) int {
// 		if a.x != b.x {
// 			return int(a.x) - int(b.x)
// 		}
// 		return int(a.y) - int(b.y)
// 	})
// 	return position
// }

// func (b *Board) adjacentKingPositions(position KingNeighbors) ([]KingNeighbors, error) {
// 	if !b.validKingPosition(position) {
// 		return nil, fmt.Errorf("invalid king position: %v", position)
// 	}
// 	unique := make(map[KingNeighbors]struct{})
// 	for i := 0; i < len(position); i++ {
// 		for j := i + 1; j < len(position); j++ {
// 			first, second := position[i], position[j]
// 			for direction := range kingDirections {
// 				third, ok := b.neighbor(first, uint8(direction))
// 				if !ok || third == second || third == position[0] || third == position[1] || third == position[2] {
// 					continue
// 				}
// 				if !b.areNeighbors(second, third) {
// 					continue
// 				}
// 				candidate := canonicalKingPosition(KingNeighbors{first, second, third})
// 				unique[candidate] = struct{}{}
// 			}
// 		}
// 	}
// 	moves := make([]KingNeighbors, 0, len(unique))
// 	for move := range unique {
// 		moves = append(moves, move)
// 	}
// 	slices.SortFunc(moves, func(a, c KingNeighbors) int {
// 		for i := range a {
// 			if a[i].x != c[i].x {
// 				return int(a[i].x) - int(c[i].x)
// 			}
// 			if a[i].y != c[i].y {
// 				return int(a[i].y) - int(c[i].y)
// 			}
// 		}
// 		return 0
// 	})
// 	return moves, nil
// }

// func (b *Board) KingMoves(position KingNeighbors, color bool) ([]KingNeighbors, error) {
// 	adjacent, err := b.adjacentKingPositions(position)
// 	if err != nil {
// 		return nil, err
// 	}
// 	moves := make([]KingNeighbors, 0, len(adjacent))
// 	for _, move := range adjacent {
// 		inCheck, checkErr := b.Check(move, color)
// 		if checkErr != nil {
// 			return nil, checkErr
// 		}
// 		if !inCheck {
// 			moves = append(moves, move)
// 		}
// 	}
// 	return moves, nil
// }

// func (b *Board) kingInCheck(position KingNeighbors, color bool, opponent KingNeighbors) (bool, error) {
// 	inCheck, err := b.Check(position, color)
// 	if err != nil || inCheck {
// 		return inCheck, err
// 	}
// 	opponentMoves, err := b.adjacentKingPositions(opponent)
// 	if err != nil {
// 		return false, err
// 	}
// 	position = canonicalKingPosition(position)
// 	return position == canonicalKingPosition(opponent) || slices.Contains(opponentMoves, position), nil
// }

// func (b *Board) KingMovesAgainst(position KingNeighbors, color bool, opponent KingNeighbors) ([]KingNeighbors, error) {
// 	moves, err := b.KingMoves(position, color)
// 	if err != nil {
// 		return nil, err
// 	}
// 	opponentMoves, err := b.adjacentKingPositions(opponent)
// 	if err != nil {
// 		return nil, err
// 	}
// 	opponent = canonicalKingPosition(opponent)
// 	safeMoves := make([]KingNeighbors, 0, len(moves))
// 	for _, move := range moves {
// 		if move != opponent && !slices.Contains(opponentMoves, move) {
// 			safeMoves = append(safeMoves, move)
// 		}
// 	}
// 	return safeMoves, nil
// }

// func (b *Board) MoveDaKing(current, next KingNeighbors, color bool) error {
// 	moves, err := b.KingMoves(current, color)
// 	if err != nil {
// 		return err
// 	}
// 	if !slices.Contains(moves, canonicalKingPosition(next)) {
// 		return fmt.Errorf("king cannot move from %v to %v", current, next)
// 	}
// 	return nil
// }

// func (b *Board) MoveKingAgainst(current, next KingNeighbors, color bool, opponent KingNeighbors) error {
// 	moves, err := b.KingMovesAgainst(current, color, opponent)
// 	if err != nil {
// 		return err
// 	}
// 	if !slices.Contains(moves, canonicalKingPosition(next)) {
// 		return fmt.Errorf("king cannot move from %v to %v", current, next)
// 	}
// 	return nil
// }

// Dummy code so it doesn't raise errors
func (b *Board) KingMovesAgainst(position KingNeighbors, color bool, opponent KingNeighbors) ([]KingNeighbors, error) {
	return []KingNeighbors{}, nil
}

func (b *Board) MoveKingAgainst(current, next KingNeighbors, color bool, opponent KingNeighbors) error {
	return nil
}

func canonicalKingPosition(position KingNeighbors) KingNeighbors {
	return KingNeighbors{}
}
