package hexchess

const (
	pScore     int16 = 1   // Each piece gets pScore points per angle it has.
	checkScore int16 = 8   // Additional score if piece is in check
	mateScore  int16 = 100 // Score when it ends in checkmate
)

// TODO hash these constants and name the .gob after the hash.

// A positive score favors the true player. Negative favors false.
func (b *Board) getScore(color bool, trueKing, falseKing KingPos) (score int16, err error) {
	score = 0
	err = nil
	// Iterate over all spaces
	for _, row := range b {
		for _, space := range row {
			if space.col {
				score += int16(len(space.ang)) * pScore
			} else {
				score -= int16(len(space.ang)) * pScore
			}
		}
	}
	// TODO count how many moves the player has

	// Add bonus points based on check:

	// true player
	isCheck, err := b.Check(trueKing, true)
	if err != nil {
		return
	}
	if isCheck {
		if color {
			score -= mateScore
		} else {
			score -= checkScore
		}
	}

	// false player
	isCheck, err = b.Check(falseKing, false)
	if err != nil {
		return
	}
	if isCheck {
		if !color {
			score += mateScore
		} else {
			score += checkScore
		}
	}
	return
}
