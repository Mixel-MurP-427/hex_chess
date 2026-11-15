package hexchess

import (
	"fmt"
	"math"
)

// Recursive decision algorithm. Uses minimax with alpha beta pruning.
func minimax(
	childMap *memory, // Memory of past choice
	pMovesMap *moveTable, // Table used by GetAllMoves function
	isOrigin bool, // true if this is the first time minimax was called.
	b *Board,
	depth uint8, // Search depth in tree
	alpha, beta int8, // used for alpha beta pruning
	who, isMaxingPlayer bool, // who is the color of the player
	maxKing, minKing KingPos,
) (int8, moveChoice, error) {

	// Check if this instance has already been computed
	strState := fmt.Sprint(state{maxKing, minKing, b})
	currentConditions := conditions{depth, isMaxingPlayer}
	computedOutcome, exists := (*childMap)[strState][currentConditions]
	if exists {
		return computedOutcome.Score, computedOutcome.Decision, nil
	}

	if isMaxingPlayer {
		// If at the bottom of the tree (the tips of the leaves)
		if depth == 0 {
			// Get score of board at this state
			score, err := b.getScore(who, maxKing, minKing)
			if err != nil {
				return 0, moveChoice{}, err
			}
			// Get positive score for maxing player
			if who {
				return score, moveChoice{}, nil
			} else {
				return score * -1, moveChoice{}, nil
			}
		}

		maxEval := int8(math.MinInt8)
		moves, err := b.GetAllMoves(maxKing, pMovesMap)
		if err != nil {
			return 0, moveChoice{}, err
		}

		choice := moveChoice{}

		// Iterate over all possible piece moves
		for from, tos := range moves {
			for _, newMove := range tos {
				for angle := range uint8(6) {
					// Store state of spaces before move
					oldFrom := b[from.X][from.Y]
					oldTo := b[newMove.X][newMove.Y]
					// Excecute the move
					b[newMove.X][newMove.Y] = rotatePiece(b[from.X][from.Y], angle)
					b[from.X][from.Y] = Piece{}

					// Call minimax on new board state
					eval, _, err := minimax(childMap, pMovesMap, false, b, depth-1, alpha, beta, !who, false, maxKing, minKing)
					if err != nil {
						return 0, moveChoice{}, err
					}

					// Move the pieces back
					b[from.X][from.Y] = oldFrom
					b[newMove.X][newMove.Y] = oldTo

					// Calculate scoring
					if eval > maxEval {
						maxEval = eval
						if isOrigin {
							choice = moveChoice{from, newMove, angle}
						}
					}
					alpha = max(alpha, eval)
					// Prune tree
					if beta <= alpha {
						break
					}
				}
			}
		}
		// Iterate over all possible king moves
		kingMoves, err := maxKing.GetLegalMoves(b, true)
		if err != nil {
			return 0, moveChoice{}, err
		}

		for _, kMove := range kingMoves {
			eval, _, err := minimax(childMap, pMovesMap, false, b, depth-1, alpha, beta, !who, false, KingPos{kMove, !maxKing.Ontop, maxKing.Col}, minKing)
			if err != nil {
				return 0, moveChoice{}, err
			}
			if eval > maxEval {
				maxEval = eval
				if isOrigin {
					choice = moveChoice{Coord{9, 9}, kMove, 9}
				}
			}
			alpha = max(alpha, eval)
			// Prune tree
			if beta <= alpha {
				break
			}
		}

		if _, exists = (*childMap)[strState]; !exists {
			(*childMap)[strState] = make(map[conditions]outcome)
		}
		(*childMap)[strState][currentConditions] = outcome{maxEval, choice}
		return maxEval, choice, nil

	} else { // When not maxingPlayer

		// If at the bottom of the tree (the tips of the leaves)
		if depth == 0 {
			// Get score of board at this state
			score, err := b.getScore(who, maxKing, minKing)
			if err != nil {
				return 0, moveChoice{}, err
			}
			// Get negative score for mining player
			if who {
				return score * -1, moveChoice{}, nil
			} else {
				return score, moveChoice{}, nil
			}
		}

		minEval := int8(math.MaxInt8)
		moves, err := b.GetAllMoves(minKing, pMovesMap)
		if err != nil {
			return 0, moveChoice{}, err
		}

		choice := moveChoice{}

		// Iterate over all possible piece moves
		for from, tos := range moves {
			for _, newMove := range tos {
				for angle := range uint8(6) {
					// Store state of spaces before move
					oldFrom := b[from.X][from.Y]
					oldTo := b[newMove.X][newMove.Y]
					// Excecute the move
					b[newMove.X][newMove.Y] = rotatePiece(b[from.X][from.Y], angle)
					b[from.X][from.Y] = Piece{}

					// Call minimax on new board state
					eval, _, err := minimax(childMap, pMovesMap, false, b, depth-1, alpha, beta, !who, true, maxKing, minKing)
					if err != nil {
						return 0, moveChoice{}, err
					}

					// Move the pieces back
					b[from.X][from.Y] = oldFrom
					b[newMove.X][newMove.Y] = oldTo

					// Calculate scoring
					if eval < minEval {
						minEval = eval
						if isOrigin {
							choice = moveChoice{from, newMove, angle}
						}
					}
					beta = min(beta, eval)
					// Prune tree
					if beta <= alpha {
						break
					}
				}
			}
		}
		// Iterate over all possible king moves
		kingMoves, err := minKing.GetLegalMoves(b, true)
		if err != nil {
			return 0, moveChoice{}, err
		}

		for _, kMove := range kingMoves {
			eval, _, err := minimax(childMap, pMovesMap, false, b, depth-1, alpha, beta, !who, true, maxKing, KingPos{kMove, !minKing.Ontop, minKing.Col})
			if err != nil {
				return 0, moveChoice{}, err
			}
			if eval < minEval {
				minEval = eval
				if isOrigin {
					choice = moveChoice{Coord{9, 9}, kMove, 9}
				}
			}
			beta = min(beta, eval)
			// Prune tree
			if beta <= alpha {
				break
			}
		}

		if _, exists = (*childMap)[strState]; !exists {
			(*childMap)[strState] = make(map[conditions]outcome)
		}
		(*childMap)[strState][currentConditions] = outcome{minEval, choice}
		return minEval, choice, nil
	}
}

// Call this one when you need a choice made
func BotChoose(depth uint8, b *Board, maxKing, minKing KingPos, parentMap *botChoices, pieceMovesMap *moveTable) (moveChoice, error) {

	// Check if this instance has already been computed
	strKey := string(depth+'0') + fmt.Sprint(state{maxKing, minKing, b})
	choice, exists := (*parentMap)[strKey]
	if exists {
		return choice, nil
	}

	familyMemory := make(memory)

	_, choice, err := minimax(&familyMemory, pieceMovesMap, true, b, depth, math.MinInt8, math.MaxInt8, maxKing.Col, true, maxKing, minKing)
	if err != nil {
		return moveChoice{}, err
	}

	// Save result to hash table
	(*parentMap)[strKey] = choice
	return choice, nil
}
