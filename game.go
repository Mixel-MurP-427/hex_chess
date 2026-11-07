package main

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

type Board [9]map[uint8]Piece

// Init empty board
func MakeBoard() Board {
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
}

func (b *Board) posExists(c Coord) bool {
	_, exists := b[c.x][c.y]
	if exists {
		return true
	}
	return false
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

	// NOTE: program bugs may occur if ang values are not ordered!

	// Iterate over shift values 0 to 5
	for i := uint8(0); i < 6; i++ {
		// Assume true until inequality is found
		output := true
		// Iterate over items in angle slices
		for j := 0; j < len(p1.ang); j++ {
			// Shift (rotate) p2 angles by 1 at a time and see if they match p1.
			if p1.ang[j] != (p2.ang[j]+i)%6 {
				output = false
				break
			}
		}
		if output == true {
			return nil
		}
	}
	return fmt.Errorf("Error: Angles do not match. p1: (%v), p2: (%v)", p1, p2)
}

// Given Coordinate and angle, finds all spaces forward.
func (b *Board) getSpacesInRow(c Coord, ang uint8) (output []Coord, err error) {
	if !b.posExists(c) {
		return nil, fmt.Errorf("Error: Invalid position. (%d, %d) doesn't exist on board", c.x, c.y)
	}

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

// Checks what pieces are in the moves path and returns available moves
func (b *Board) getLegalMoves(moves []Coord, myCol bool) []Coord {
	for i, move := range moves {
		if len(b[move.x][move.y].ang) != 0 {
			if b[move.x][move.y].col == myCol {
				return moves[:i]
			} else {
				return moves[:i+1]
			}
		}
	}
	return moves
}

// Accepts Coord of piece before and after, the movement angle, and the final Piece state.
func (b *Board) Move(c1 Coord, mAng uint8, c2 Coord, p2 Piece) error {

	// Error checking

	// Check validity of coordinates
	if !b.posExists(c1) {
		return fmt.Errorf("Error: Invalid position. c1 (%d, %d) doesn't exist on board", c1.x, c1.y)
	}
	if !b.posExists(c2) {
		return fmt.Errorf("Error: Invalid position. c2 (%d, %d) doesn't exist on board", c2.x, c2.y)
	}

	p1 := b[c1.x][c1.y]

	// Check if piece moved matches original piece
	err := piecesMatch(p1, p2)
	if err != nil {
		return err
	}

	// check if angle of movement is valid
	if !slices.Contains(p1.ang, mAng) {
		return nil
	}

	// Search for valid moves

	moves, err := b.getSpacesInRow(c1, mAng)
	if err != nil {
		return err
	}
	moves = b.getLegalMoves(moves, p2.col)

	if !slices.Contains(moves, c2) {
		return fmt.Errorf("Error: Invalid move!")
	}

	// Set board with new move
	b[c1.x][c1.y] = Piece{}
	b[c2.x][c2.y] = p2

	return nil
}

// Code for testing
func main() {
	fmt.Println("Hello!")

	board := MakeBoard()

	board[7][1] = Piece{true, []uint8{33}}

	moves, err := board.getSpacesInRow(Coord{4, 4}, 0)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(board.getLegalMoves(moves, true))

}
