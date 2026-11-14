package hexchess

// Get available moves that each piece can make
func (b *Board) GetAllMoves(theKing KingPos) (finalMoves map[Coord][]Coord, err error) {
	finalMoves = make(map[Coord][]Coord)

	// Iterate over all spaces on the board
	for xInt, row := range b {
		x := uint8(xInt)
		for y, piece := range row {

			// If the player is not allowed to manipulate this piece
			if piece.Ang == 0 || piece.Col != theKing.Col {
				continue
			}
			// Get legal moves for this piece
			moves, err := b.GetLegalMoves(Coord{x, y}, true, theKing)
			if err != nil {
				return map[Coord][]Coord{}, err
			}
			if len(moves) != 0 {
				finalMoves[Coord{x, y}] = moves
			}
		}
	}
	return finalMoves, nil
}
