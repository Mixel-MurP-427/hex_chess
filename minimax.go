package hexchess

import (
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"strings"
)

type moveTable map[string]map[Coord][]Coord

// Get available moves that each piece can make
func (b *Board) GetAllMoves(theKing KingPos, saveTable *moveTable) (finalMoves map[Coord][]Coord, err error) {

	// Check if this instance has already been computed
	strKey := fmt.Sprint(state{theKing, KingPos{}, b})
	allMoves, exists := (*saveTable)[strKey]
	if exists {
		return allMoves, nil
	}

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
	(*saveTable)[strKey] = finalMoves
	return finalMoves, nil
}

// Setup for choice memory

type state struct {
	trueKing  KingPos
	falseKing KingPos
	board     *Board
}

type moveChoice struct {
	FromCoord Coord // If == {9, 9}, then the next coord is a KingPos.KCoord
	ToCoord   Coord
	Rot       uint8
}

type botChoices map[string]moveChoice

type conditions struct {
	SearchDepth uint8
	IsMaxPlyr   bool
}

type outcome struct {
	Score    int8
	Decision moveChoice
}

type memory map[string]map[conditions]outcome

func (board *Board) String() string {

	var keys = [9][]uint8{
		{4, 5, 6, 7, 8},
		{3, 4, 5, 6, 7, 8},
		{2, 3, 4, 5, 6, 7, 8},
		{1, 2, 3, 4, 5, 6, 7, 8},
		{0, 1, 2, 3, 4, 5, 6, 7, 8},
		{0, 1, 2, 3, 4, 5, 6, 7},
		{0, 1, 2, 3, 4, 5, 6},
		{0, 1, 2, 3, 4, 5},
		{0, 1, 2, 3, 4},
	}

	var b strings.Builder
	b.Grow(144)
	b.WriteString("b[")

	for x := range keys {
		b.WriteString("[")

		for _, y := range keys[x] {

			if board[x][y].Ang == 0 {
				b.WriteString("  ")

			} else {

				if board[x][y].Col {
					b.WriteByte('!')
				} else {
					b.WriteByte('"')
				}
				b.WriteByte('#' + board[x][y].Ang)

			}
		}
		b.WriteByte(']')
	}
	b.WriteByte(']')

	return b.String()
}

func (k KingPos) String() string {

	var b strings.Builder
	b.Grow(8)

	b.WriteString("K{")

	b.WriteByte('0' + k.KCoord.X)
	b.WriteByte('0' + k.KCoord.Y)

	if k.Ontop {
		if k.Col {
			b.WriteByte('T')
		} else {
			b.WriteByte('F')
		}
	} else {
		if k.Col {
			b.WriteByte('t')
		} else {
			b.WriteByte('f')
		}
	}

	b.WriteByte('}')

	return b.String()
}

func (s state) String() string {

	var b strings.Builder
	b.Grow(160)

	b.WriteString(fmt.Sprint(s.trueKing))
	b.WriteString(fmt.Sprint(s.falseKing))
	b.WriteString(fmt.Sprint(s.board))

	return b.String()
}

func loadBotChoices(path string) (botChoices, error) {
	var loaded = make(botChoices)
	readFile, err := os.Open(path)
	if err != nil {
		return loaded, err
	}
	decoder := gob.NewDecoder(readFile)
	if err = decoder.Decode(&loaded); err != nil {
		return nil, err
	}
	if err := readFile.Close(); err != nil {
		return nil, err
	}
	return loaded, nil
}

func (data *botChoices) saveBotChoices(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(file)
	if err = encoder.Encode(*data); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return nil
}

const moveTablePath = "hash_tables/GetAllMovesTable.go"

func loadMoveTable(path string) (moveTable, error) {
	var loaded = make(moveTable)
	readFile, err := os.Open(path)
	if err != nil {
		return loaded, err
	}
	decoder := gob.NewDecoder(readFile)
	if err = decoder.Decode(&loaded); err != nil {
		return nil, err
	}
	if err := readFile.Close(); err != nil {
		return nil, err
	}
	return loaded, nil
}

func (data *moveTable) saveMoveTable(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(file)
	if err = encoder.Encode(*data); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return nil
}

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
