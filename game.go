package hexchess

import (
	"fmt"
	"math/bits"
	"slices"
)

type Coord struct {
	x uint8
	y uint8
}
type Piece struct {
	col bool  // Color: true represents self, false represents opponent.
	ang uint8 // Angle: represents a list of direction values 0 to 5, where each bit tells whether that direction is in the list. Bits 7 and 8 are ignored.
}

// If ontop, the king sits directly on top of coord. Else, it directly below coord.
type KingPos struct {
	coord Coord
	ontop bool
}
type KingNeighbors [3]Coord

var (
	tKStart = KingPos{Coord{2, 2}, true}
	fKStart = KingPos{Coord{6, 6}, false}
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
			{4: Piece{true, 18}, 5: Piece{true, 4}, 6: Piece{}, 7: Piece{}, 8: Piece{}},
			{3: Piece{true, 42}, 4: Piece{true, 4}, 5: Piece{}, 6: Piece{}, 7: Piece{}, 8: Piece{}},
			{2: Piece{true, 6}, 3: Piece{true, 4}, 4: Piece{}, 5: Piece{}, 6: Piece{}, 7: Piece{}, 8: Piece{}},
			{1: Piece{true, 63}, 2: Piece{true, 2}, 3: Piece{}, 4: Piece{}, 5: Piece{}, 6: Piece{}, 7: Piece{}, 8: Piece{false, 16}},
			{0: Piece{true, 36}, 1: Piece{true, 2}, 2: Piece{}, 3: Piece{}, 4: Piece{}, 5: Piece{}, 6: Piece{}, 7: Piece{false, 16}, 8: Piece{false, 36}},
			{0: Piece{true, 2}, 1: Piece{}, 2: Piece{}, 3: Piece{}, 4: Piece{}, 5: Piece{}, 6: Piece{false, 16}, 7: Piece{false, 42}},
			{0: Piece{}, 1: Piece{}, 2: Piece{}, 3: Piece{}, 4: Piece{}, 5: Piece{false, 32}, 6: Piece{false, 48}},
			{0: Piece{}, 1: Piece{}, 2: Piece{}, 3: Piece{}, 4: Piece{false, 32}, 5: Piece{false, 63}},
			{0: Piece{}, 1: Piece{}, 2: Piece{}, 3: Piece{false, 32}, 4: Piece{false, 18}},
		}
	}
}

func (board *Board) clone() Board {
	var cloned Board
	for x, row := range board {
		cloned[x] = make(map[uint8]Piece, len(row))
		for y, piece := range row {
			cloned[x][y] = piece
		}
	}
	return cloned
}

// TODO change so that it compares numbers instead of checking the actual board. This will faulty additions to the board.
func posExists(c Coord) bool {
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

func rotatePiece(piece Piece, steps int) Piece {
	return Piece{piece.col, bits.RotateLeft8(piece.ang, steps)}
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
		if posExists(Coord{i, j}) {
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
	for q := range uint8(6) {
		// If angle q is not one of the angles
		if (p.ang>>q)&1 == 0 {
			continue
		}
		pFound := false
		moves, gErr := b.getSpacesInRow(c, q)
		if gErr != nil {
			return nil, gErr
		}
		// Iterate over moves
		for i, move := range moves {
			// If the space is not empty
			if b[move.x][move.y].ang != 0 {
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
func (b *Board) Move(c1 Coord, c2 Coord, rotation int) error {

	if rotation > 5 {
		return fmt.Errorf("invalid rotation %d: must be between 0 and 5", rotation)
	}

	moves, err := b.getLegalMoves(c1)
	if err != nil {
		return err
	}

	if !slices.Contains(moves, c2) {
		return fmt.Errorf("Error: Position (%d, %d) is not in the list of legal moves: %v", c2.x, c2.y, moves)
	}

	// Set board with new move
	b[c2.x][c2.y] = rotatePiece(b[c1.x][c1.y], rotation)
	b[c1.x][c1.y] = Piece{}

	return nil
}

// Return whether or not the king is in check
func (b *Board) Check(kN KingNeighbors, kCol bool) (bool, error) {
	// Trace from the king outward to find if threats are in his path.

	// Iterate over the king's neighboring spaces
	for _, space := range kN {
		// Continue if an enemy piece sits in this space.
		if b[space.x][space.y].ang != 0 && b[space.x][space.y].col != kCol {
			continue
		}
		// Iterate over all six directions
		for a := uint8(0); a < 6; a++ {
			// Continue if the adjacent space does not also border the king.
			switch a {
			case 0:
				if !slices.Contains(kN[:], Coord{space.x + 1, space.y - 1}) {
					continue
				}
			case 1:
				if !slices.Contains(kN[:], Coord{space.x + 1, space.y}) {
					continue
				}
			case 2:
				if !slices.Contains(kN[:], Coord{space.x, space.y + 1}) {
					continue
				}
			case 3:
				if !slices.Contains(kN[:], Coord{space.x - 1, space.y + 1}) {
					continue
				}
			case 4:
				if !slices.Contains(kN[:], Coord{space.x - 1, space.y}) {
					continue
				}
			case 5:
				if !slices.Contains(kN[:], Coord{space.x, space.y - 1}) {
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
				if p.ang != 0 {
					// If it is an enemy piece pointing at the king
					if p.col != kCol && (p.ang>>((a+3)%6))&1 == 1 { // get the ((a+3)%6)th bit (opposite direction of var a) and check if == 1
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

func (k *KingPos) getKingNeighbors() KingNeighbors {
	if k.ontop {
		return KingNeighbors{k.coord, Coord{k.coord.x, k.coord.y + 1}, Coord{k.coord.x + 1, k.coord.y}}
	} else {
		return KingNeighbors{k.coord, Coord{k.coord.x, k.coord.y - 1}, Coord{k.coord.x - 1, k.coord.y}}
	}
}

// Determines the king position is valid
func (n *KingNeighbors) validKingNeighbors() bool {
	if posExists(n[0]) && posExists(n[1]) && posExists(n[2]) {
		return true
	} else {
		return false
	}
}

// Get the locations where the king can move on the board
func (k *KingPos) getMoves() (moves []Coord) {
	// Get available adjacent moves
	if k.ontop {
		moves = []Coord{{k.coord.x, k.coord.y + 1}, {k.coord.x + 1, k.coord.y}, {k.coord.x + 1, k.coord.y + 1}}
	} else {
		moves = []Coord{{k.coord.x, k.coord.y - 1}, {k.coord.x - 1, k.coord.y}, {k.coord.x - 1, k.coord.y - 1}}
	}

	// Validate that these moves lie on the board
	keepIndex := 0
	// Sort moves and only return valid ones
	for _, move := range moves {
		newK := KingPos{move, !k.ontop}
		newN := newK.getKingNeighbors()
		if newN.validKingNeighbors() {
			moves[keepIndex] = move
			keepIndex++
		}
	}
	return moves[:keepIndex]
}

func (k1 *KingPos) Move(b *Board, c2 Coord, kCol bool) error {
	// Get neighbors and validate
	n1 := k1.getKingNeighbors()
	if !n1.validKingNeighbors() {
		return fmt.Errorf("Error: Invalid king location: (%v) not on board.", k1)
	}
	k2 := KingPos{c2, !k1.ontop}
	n2 := k2.getKingNeighbors()
	if !n2.validKingNeighbors() {
		return fmt.Errorf("Error: Invalid king location: (%v) not on board.", c2)
	}

	// Check if this move is to an adjacent triangle
	if !slices.Contains(k1.getMoves(), c2) {
		return fmt.Errorf("Error: Invalid king move from (%v) to (%v). Move must be one of the following: %v", k1, c2, k1.getMoves())
	}

	isCheck, err := b.Check(n2, kCol)
	if err != nil {
		return err
	}
	if isCheck {
		return fmt.Errorf("Cannot move king at (%v) to (%v) because it would result in check.", k1, k2)
	}
	// Move the king
	*k1 = k2

	return nil
}
