package hexchess

import (
	"fmt"
	"slices"
)

type Coord struct {
	X uint8
	Y uint8
}
type Piece struct {
	Col bool  // Color: true represents self, false represents opponent.
	Ang uint8 // Angle: represents a list of direction values 0 to 5, where each bit tells whether that direction is in the list. Bits 7 and 8 are ignored.
}

// If ontop, the king sits directly on top of coord. Else, it directly below coord.
type KingPos struct {
	KCoord Coord
	Ontop  bool
}
type KingNeighbors [3]Coord

var (
	TheTrueKing  = KingPos{Coord{2, 2}, true}
	TheFalseKing = KingPos{Coord{6, 6}, false}
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

func posExists(c Coord) bool {
	// Must be between 0 and 8
	if c.X >= 9 || c.Y >= 9 {
		return false
	}
	if c.X <= 4 {
		if c.Y >= 4-c.X {
			return true
		}
	} else {
		if c.Y <= 12-c.X {
			return true
		}
	}
	return false
}

func rotatePiece(piece Piece, steps uint8) Piece {
	var newAng uint8
	// Normalize rotation amount to be within 0-5
	steps %= 6
	if steps == 0 {
		newAng = piece.Ang
	}

	// Perform the 6-bit left rotation and re-apply the mask
	newAng = ((piece.Ang << steps) | (piece.Ang >> uint(6-steps))) & 0x3F
	return Piece{piece.Col, newAng}
}

// Given Coordinate and angle, finds all spaces forward.
func (b *Board) getSpacesInRow(c Coord, ang uint8) (output []Coord, err error) {

	var xs, ys []uint8

	switch ang {
	case 0:
		// x increases, y decreases
		for i, j := c.X+1, c.Y-1; i < 9 && j < 9; i, j = i+1, j-1 {
			xs = append(xs, i)
			ys = append(ys, j)
		}
	case 1:
		// x increases, y stays constant
		for i := c.X + 1; i < 9; i++ {
			xs = append(xs, i)
			ys = append(ys, c.Y)
		}
	case 2:
		// x stays constant, y increases
		for j := c.Y + 1; j < 9; j++ {
			xs = append(xs, c.X)
			ys = append(ys, j)
		}
	case 3:
		// x decreases, y increases
		for i, j := c.X-1, c.Y+1; i < 9 && j < 9; i, j = i-1, j+1 {
			xs = append(xs, i)
			ys = append(ys, j)
		}
	case 4:
		// x decreases, y stays constant
		for i := c.X - 1; i < 9; i-- {
			xs = append(xs, i)
			ys = append(ys, c.Y)
		}
	case 5:
		// x stays constant, y decreases
		for j := c.Y - 1; j < 9; j-- {
			xs = append(xs, c.X)
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
func (b *Board) GetLegalMoves(c Coord, check4Check bool) (output []Coord, err error) {
	output = []Coord{}
	p, ok := b[c.X][c.Y]
	if !ok {
		return nil, fmt.Errorf("Error: Invalid position. (%d, %d) doesn't exist on board", c.X, c.Y)
	}

	// Iterate over available angle of the piece.
	for q := range uint8(6) {
		// If angle q is not one of the angles
		if (p.Ang>>q)&1 == 0 {
			continue
		}
		moves, gErr := b.getSpacesInRow(c, q)
		if gErr != nil {
			return nil, gErr
		}
		// Iterate over moves
		for i, move := range moves {
			// If the space is not empty
			if b[move.X][move.Y].Ang != 0 {
				// If the colors match
				if b[move.X][move.Y].Col == p.Col {
					moves = moves[:i]
				} else {
					moves = moves[:i+1]
				}
				break
			}
		}

		// Check if each move would put the moving player in check
		if check4Check {

			// Get neighbors of king
			kCol := b[c.X][c.Y].Col
			var n KingNeighbors
			if kCol {
				n = TheTrueKing.getKingNeighbors()
			} else {
				n = TheFalseKing.getKingNeighbors()
			}

			keepIndex := 0
			// Iterate over moves
			for _, move := range moves {

				// Clone board
				bClone := b.clone()
				// move piece
				bClone[move.X][move.Y] = b[c.X][c.Y]

				isCheck, err := bClone.Check(n, kCol)
				if err != nil {
					return []Coord{}, err
				}
				// Exclude moves that result in check
				if isCheck {
					moves[keepIndex] = move
					keepIndex++
				}
			}
			// Add available moves in this direction to output
			output = append(output, moves[:keepIndex]...)
		} else {
			output = append(output, moves...)
		}
	}
	return output, nil
}

// Accepts Coord of piece before and after, the movement angle, and the final Piece state.
func (b *Board) Move(c1 Coord, c2 Coord, rotation uint8) error {

	if rotation > 5 {
		return fmt.Errorf("invalid rotation %d: must be between 0 and 5", rotation)
	}

	moves, err := b.GetLegalMoves(c1, false)
	if err != nil {
		return err
	}

	if !slices.Contains(moves, c2) {
		return fmt.Errorf("Error: Position (%d, %d) is not in the list of legal moves: %v", c2.X, c2.Y, moves)
	}

	// Check for check
	kCol := b[c1.X][c1.Y].Col
	var n KingNeighbors
	if kCol {
		n = TheTrueKing.getKingNeighbors()
	} else {
		n = TheFalseKing.getKingNeighbors()
	}

	isCheck, err := b.Check(n, kCol)
	if err != nil {
		return err
	}
	if isCheck {
		return fmt.Errorf("Cannot move from (%v) to (%v) because it would result in check.", c1, c2)
	}

	// Set board with new move
	b[c2.X][c2.Y] = rotatePiece(b[c1.X][c1.Y], rotation)
	b[c1.X][c1.Y] = Piece{}

	return nil
}

// Return whether or not the king is in check
func (b *Board) Check(kN KingNeighbors, kCol bool) (bool, error) {
	// Trace from the king outward to find if threats are in his path.

	// Iterate over the king's neighboring spaces
	for _, space := range kN {
		// Continue if an enemy piece sits in this space.
		if b[space.X][space.Y].Ang != 0 && b[space.X][space.Y].Col != kCol {
			continue
		}
		// Iterate over all six directions
		for a := uint8(0); a < 6; a++ {
			// Continue if the adjacent space does not also border the king.
			switch a {
			case 0:
				if !slices.Contains(kN[:], Coord{space.X + 1, space.Y - 1}) {
					continue
				}
			case 1:
				if !slices.Contains(kN[:], Coord{space.X + 1, space.Y}) {
					continue
				}
			case 2:
				if !slices.Contains(kN[:], Coord{space.X, space.Y + 1}) {
					continue
				}
			case 3:
				if !slices.Contains(kN[:], Coord{space.X - 1, space.Y + 1}) {
					continue
				}
			case 4:
				if !slices.Contains(kN[:], Coord{space.X - 1, space.Y}) {
					continue
				}
			case 5:
				if !slices.Contains(kN[:], Coord{space.X, space.Y - 1}) {
					continue
				}
			}
			moves, err := b.getSpacesInRow(space, a)
			if err != nil {
				return false, err
			}
			for _, move := range moves {
				p := b[move.X][move.Y]
				// If there is a piece here
				if p.Ang != 0 {
					// If it is an enemy piece pointing at the king
					if p.Col != kCol && (p.Ang>>((a+3)%6))&1 == 1 { // get the ((a+3)%6)th bit (opposite direction of var a) and check if == 1
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
	if k.Ontop {
		return KingNeighbors{k.KCoord, Coord{k.KCoord.X, k.KCoord.Y + 1}, Coord{k.KCoord.X + 1, k.KCoord.Y}}
	} else {
		return KingNeighbors{k.KCoord, Coord{k.KCoord.X, k.KCoord.Y - 1}, Coord{k.KCoord.X - 1, k.KCoord.Y}}
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
func (k *KingPos) GetLegalMoves(b *Board, check4Check bool) (moves []Coord, err error) {
	// Get available adjacent moves
	if k.Ontop {
		moves = []Coord{{k.KCoord.X, k.KCoord.Y + 1}, {k.KCoord.X + 1, k.KCoord.Y}, {k.KCoord.X + 1, k.KCoord.Y + 1}}
	} else {
		moves = []Coord{{k.KCoord.X, k.KCoord.Y - 1}, {k.KCoord.X - 1, k.KCoord.Y}, {k.KCoord.X - 1, k.KCoord.Y - 1}}
	}

	// Validate that these moves lie on the board
	keepIndex := 0
	// Sort moves and only return valid ones
	for _, move := range moves {
		newK := KingPos{move, !k.Ontop}
		newN := newK.getKingNeighbors()
		// if this move lies on the board
		if newN.validKingNeighbors() {
			// if move results in check
			if check4Check {
				isCheck, err := b.Check(newN, !k.Ontop)
				if err != nil {
					return []Coord{}, err
				}
				if isCheck {
					continue
				}
			}
			moves[keepIndex] = move
			keepIndex++
		}
	}
	return moves[:keepIndex], nil
}

func (k1 *KingPos) Move(b *Board, c2 Coord, kCol bool) error {
	// Get neighbors and validate
	n1 := k1.getKingNeighbors()
	if !n1.validKingNeighbors() {
		return fmt.Errorf("Error: Invalid king location: (%v) not on board.", k1)
	}
	k2 := KingPos{c2, !k1.Ontop}
	n2 := k2.getKingNeighbors()
	if !n2.validKingNeighbors() {
		return fmt.Errorf("Error: Invalid king location: (%v) not on board.", c2)
	}

	// Check if this move is to an adjacent triangle
	moves, err := k1.GetLegalMoves(b, false)
	if err != nil {
		return err
	}
	if !slices.Contains(moves, c2) {
		return fmt.Errorf("Error: Invalid king move from (%v) to (%v). Move must be one of the following: %v", k1, c2, moves)
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
