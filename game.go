package hexchess

import (
	"fmt"
	"slices"
)

type Coord struct {
	x uint8
	y uint8
}
type Piece struct {
	col bool    // Color: true represents self, false represents opponent.
	ang []uint8 // Angle: Slice of any values 0 to 5. Must be in order.
}

type KingPos [3]Coord

var (
	tKStart = KingPos{Coord{2, 2}, Coord{3, 2}, Coord{2, 3}}
	fKStart = KingPos{Coord{6, 6}, Coord{5, 6}, Coord{6, 5}}
)

type Board [9]map[uint8]Piece

// Init board. True for empty, false for setup.
func SetupBoard(empty bool) Board {
	if empty {
		return Board{
			{4: Piece{}, 5: Piece{}, 6: Piece{}, 7: Piece{}, 8: Piece{}},
			{3: Piece{}, 4: Piece{}, 5: Piece{}, 6: Piece{}, 7: Piece{}, 8: Piece{}},
			{2: Piece{}, 3: Piece{}, 4: Piece{}, 5: Piece{}, 6: Piece{}, 7: Piece{}, 8: Piece{}},
			{1: Piece{}, 2: Piece{}, 3: Piece{}, 4: Piece{}, 5: Piece{}, 6: Piece{}, 7: Piece{}, 8: Piece{}},
			{0: Piece{}, 1: Piece{}, 2: Piece{}, 3: Piece{}, 4: Piece{}, 5: Piece{}, 6: Piece{}, 7: Piece{}, 8: Piece{}},
			{0: Piece{}, 1: Piece{}, 2: Piece{}, 3: Piece{}, 4: Piece{}, 5: Piece{}, 6: Piece{}, 7: Piece{}},
			{0: Piece{}, 1: Piece{}, 2: Piece{}, 3: Piece{}, 4: Piece{}, 5: Piece{}, 6: Piece{}},
			{0: Piece{}, 1: Piece{}, 2: Piece{}, 3: Piece{}, 4: Piece{}, 5: Piece{}},
			{0: Piece{}, 1: Piece{}, 2: Piece{}, 3: Piece{}, 4: Piece{}},
		}
	} else {
		return Board{
			{4: Piece{true, []uint8{1, 4}}, 5: Piece{true, []uint8{2}}, 6: Piece{}, 7: Piece{}, 8: Piece{}},
			{3: Piece{true, []uint8{1, 3, 5}}, 4: Piece{true, []uint8{2}}, 5: Piece{}, 6: Piece{}, 7: Piece{}, 8: Piece{}},
			{2: Piece{true, []uint8{1, 2}}, 3: Piece{true, []uint8{2}}, 4: Piece{}, 5: Piece{}, 6: Piece{}, 7: Piece{}, 8: Piece{}},
			{1: Piece{true, []uint8{0, 1, 2, 3, 4, 5}}, 2: Piece{true, []uint8{1}}, 3: Piece{}, 4: Piece{}, 5: Piece{}, 6: Piece{}, 7: Piece{}, 8: Piece{false, []uint8{4}}},
			{0: Piece{true, []uint8{2, 5}}, 1: Piece{true, []uint8{1}}, 2: Piece{}, 3: Piece{}, 4: Piece{}, 5: Piece{}, 6: Piece{}, 7: Piece{false, []uint8{4}}, 8: Piece{false, []uint8{2, 5}}},
			{0: Piece{true, []uint8{1}}, 1: Piece{}, 2: Piece{}, 3: Piece{}, 4: Piece{}, 5: Piece{}, 6: Piece{false, []uint8{4}}, 7: Piece{false, []uint8{1, 3, 5}}},
			{0: Piece{}, 1: Piece{}, 2: Piece{}, 3: Piece{}, 4: Piece{}, 5: Piece{false, []uint8{5}}, 6: Piece{false, []uint8{4, 5}}},
			{0: Piece{}, 1: Piece{}, 2: Piece{}, 3: Piece{}, 4: Piece{false, []uint8{5}}, 5: Piece{false, []uint8{0, 1, 2, 3, 4, 5}}},
			{0: Piece{}, 1: Piece{}, 2: Piece{}, 3: Piece{false, []uint8{5}}, 4: Piece{false, []uint8{1, 4}}},
		}
	}
}

// TODO change so that it compares numbers instead of checking the actual board. This will faulty additions to the board.
func (b *Board) posExists(c Coord) bool {
	// Must be between 0 and 8
	if c.x >= 9 || c.y >= 9 {
		return false
	}
	if c.x <= 4 {
		if c.y >= 4-c.x {
			return true
		}
	} else {
		if c.y <= 12-c.x {
			return true
		}
	}
	// _, exists := b[c.x][c.y]
	// if exists {
	// 	return true
	// }
	return false
}

func validateAngles(angles []uint8) error {
	for i, angle := range angles {
		if angle > 5 {
			return fmt.Errorf("invalid angle in %v: angles must be between 0 and 5", angles)
		}
		if i > 0 && angles[i-1] > angle {
			return fmt.Errorf("invalid angles %v: angles must be unique and sorted", angles)
		}
	}
	return nil
}

func rotatePiece(piece Piece, steps uint8) Piece {
	rotated := Piece{col: piece.col, ang: make([]uint8, len(piece.ang))}
	for i, angle := range piece.ang {
		rotated.ang[i] = (angle + steps) % 6
	}
	slices.Sort(rotated.ang)
	return rotated
}

func piecesMatch(p1, p2 Piece) error {
	// Return error if either piece is empty
	if len(p1.ang) == 0 {
		return fmt.Errorf("Error: Piece p1 (%v) cannot be empty.", p1)
	}
	if len(p2.ang) == 0 {
		return fmt.Errorf("Error: Piece p2 (%v) cannot be empty.", p2)
	}
	// Return error if colors do not match.
	if p1.col != p2.col {
		return fmt.Errorf("Error: Pieces do not have matching color: (%t), (%t)", p1.col, p2.col)
	}
	// Return error if number of angles do not match
	if len(p1.ang) != len(p2.ang) {
		return fmt.Errorf("Error: Angles do not match. p1: (%v), p2: (%v)", p1, p2)
	}
	if err := validateAngles(p1.ang); err != nil {
		return fmt.Errorf("invalid p1 angles: %w", err)
	}
	if err := validateAngles(p2.ang); err != nil {
		return fmt.Errorf("invalid p2 angles: %w", err)
	}

	// Iterate over shift values 0 to 5
	for i := uint8(0); i < 6; i++ {
		rotated := rotatePiece(p2, i)
		if slices.Equal(p1.ang, rotated.ang) {
			return nil
		}
	}
	return fmt.Errorf("Error: Angles do not match. p1: (%v), p2: (%v)", p1, p2)
}

// Given Coordinate and angle, finds all spaces forward.
func (b *Board) getSpacesInRow(c Coord, ang uint8) (output []Coord, err error) {

	var xs, ys []uint8

	switch ang {
	case 0:
		// x increases, y decreases
		for i, j := c.x+1, c.y-1; i < 9 && j < 9; i, j = i+1, j-1 {
			xs = append(xs, i)
			ys = append(ys, j)
		}
	case 1:
		// x increases, y stays constant
		for i := c.x + 1; i < 9; i++ {
			xs = append(xs, i)
			ys = append(ys, c.y)
		}
	case 2:
		// x stays constant, y increases
		for j := c.y + 1; j < 9; j++ {
			xs = append(xs, c.x)
			ys = append(ys, j)
		}
	case 3:
		// x decreases, y increases
		for i, j := c.x-1, c.y+1; i < 9 && j < 9; i, j = i-1, j+1 {
			xs = append(xs, i)
			ys = append(ys, j)
		}
	case 4:
		// x decreases, y stays constant
		for i := c.x - 1; i < 9; i-- {
			xs = append(xs, i)
			ys = append(ys, c.y)
		}
	case 5:
		// x stays constant, y decreases
		for j := c.y - 1; j < 9; j-- {
			xs = append(xs, c.x)
			ys = append(ys, j)
		}
	default:
		return nil, fmt.Errorf("Error: Invalid angle of direction: %d. Must be between 0 and 5", ang)
	}

	// Filter or append valid positions
	for idx := range xs {
		i, j := xs[idx], ys[idx]
		if b.posExists(Coord{i, j}) {
			output = append(output, Coord{i, j})
		}
	}

	return output, nil
}

// Accepts Coord and angle direction
func (b *Board) getLegalMoves(c Coord) (output []Coord, err error) {
	output = []Coord{}
	p, ok := b[c.x][c.y]
	if !ok {
		return nil, fmt.Errorf("Error: Invalid position. (%d, %d) doesn't exist on board", c.x, c.y)
	}

	// Iterate over available angle of the piece.
	for _, a := range p.ang {
		pFound := false
		moves, gErr := b.getSpacesInRow(c, a)
		if gErr != nil {
			return nil, gErr
		}
		// Iterate over moves
		for i, move := range moves {
			// If the space is not empty
			if len(b[move.x][move.y].ang) != 0 {
				pFound = true
				// If the colors match
				if b[move.x][move.y].col == p.col {
					output = append(output, moves[:i]...)
				} else {
					output = append(output, moves[:i+1]...)
				}
				break
			}
		}
		// If the entire move path was empty
		if !pFound {
			output = append(output, moves...)
		}
	}
	return output, nil
}

// Accepts Coord of piece before and after, the movement angle, and the final Piece state.
// TODO instead of inputing p2, input a rotation amount 0-5
func (b *Board) Move(c1 Coord, c2 Coord, p2 Piece) error {

	moves, err := b.getLegalMoves(c1)
	if err != nil {
		return err
	}

	if !slices.Contains(moves, c2) {
		return fmt.Errorf("Error: Position (%d, %d) is not in the list of legal moves: %v", c2.x, c2.y, moves)
	}
	// Check if piece moved matches original piece
	err = piecesMatch(b[c1.x][c1.y], p2)
	if err != nil {
		return err
	}

	// Set board with new move
	b[c1.x][c1.y] = Piece{}
	b[c2.x][c2.y] = p2

	return nil
}

// Return whether or not the king is in check
func (b *Board) Check(kPos KingPos, kCol bool) (bool, error) {
	// Trace from the king outward to find if threats are in his path.

	// Iterate over the king's neighboring spaces
	for _, space := range kPos {
		// Continue if an enemy piece sits in this space.
		if len(b[space.x][space.y].ang) != 0 && b[space.x][space.y].col != kCol {
			continue
		}
		// Iterate over all six directions
		for a := uint8(0); a < 6; a++ {
			// Continue if the adjacent space does not also border the king.
			switch a {
			case 0:
				if !slices.Contains(kPos[:], Coord{space.x + 1, space.y - 1}) {
					continue
				}
			case 1:
				if !slices.Contains(kPos[:], Coord{space.x + 1, space.y}) {
					continue
				}
			case 2:
				if !slices.Contains(kPos[:], Coord{space.x, space.y + 1}) {
					continue
				}
			case 3:
				if !slices.Contains(kPos[:], Coord{space.x - 1, space.y + 1}) {
					continue
				}
			case 4:
				if !slices.Contains(kPos[:], Coord{space.x - 1, space.y}) {
					continue
				}
			case 5:
				if !slices.Contains(kPos[:], Coord{space.x, space.y - 1}) {
					continue
				}
			}
			moves, err := b.getSpacesInRow(space, a)
			if err != nil {
				return false, err
			}
			for _, move := range moves {
				p := b[move.x][move.y]
				// If there is a piece here
				if len(p.ang) != 0 {
					// If it is an enemy piece pointing at the king
					if p.col != kCol && slices.Contains(p.ang, (a+3)%6) {
						// Check situation found!
						return true, nil
					}
					// Do not check the pieces behind this one.
					break
				}
			}
		}
	}
	return false, nil
}

func (k *KingPos) MoveKing(newC Coord) error {
	return nil
}
