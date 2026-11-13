package hexchess

import "math/bits"

const (
	pScore     int16 = 1   // Each piece gets pScore points per angle it has.
	checkScore int16 = 8   // Additional score if piece is in check
	mateScore  int16 = 100 // Score when it ends in checkmate
)

// TODO hash these constants and name the .gob after the hash.

// A positive score favors the true player. Negative favors false.
func (b *Board) getScore(color bool, king1, king2 KingPos) (score int16, err error) {
	score = 0
	err = nil
	// Iterate over all spaces
	for _, row := range b {
		for _, space := range row {
			// Add to score based on how many angles the piece has
			if space.Col {
				score += int16(bits.OnesCount8(space.Ang)) * pScore
			} else {
				score -= int16(bits.OnesCount8(space.Ang)) * pScore
			}
		}
	}
	// TODO count how many moves the player has

	// Add bonus points based on check:

	var isCheck bool
	for _, king := range [2]KingPos{king1, king2} {

		isCheck, err = b.Check(king)
		if err != nil {
			return
		}
		if isCheck {
			// add or subtract based on color
			add := int16(1)
			if king.Col {
				add = -1
			}

			if color == king.Col {
				score += mateScore * add
			} else {
				score += checkScore * add
			}
		}
	}
	return
}
