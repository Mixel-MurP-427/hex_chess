package hexchess

import (
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

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

// Setup for choice memory

type state struct {
	board     Board
	trueKing  KingPos
	falseKing KingPos
}

type moveChoice struct {
	FromCoord Coord // If == {9, 9}, then the next coord is a KingPos.KCoord
	ToCoord   Coord
	Rot       uint8
}

type conditions struct {
	SearchDepth uint8
	IsMaxPlyr   bool
}

type outcome struct {
	Score    int8
	Decision moveChoice
}

type memory map[string]map[conditions]outcome

// Stringer. Call fmt.Sprint(state) to convert it to a string.
func (s state) String() string {
	board := s.board
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

	totalEntries := 0
	for i := range board {
		totalEntries += len(board[i])
	}
	b.Grow(56 + totalEntries*24 + 64) // TODO fix this
	b.WriteString("s{b:[")
	var number [3]byte

	for x := range keys {
		if x > 0 {
			b.WriteByte(' ')
		}
		b.WriteString("m[")

		first := true
		for _, y := range keys[x] {
			if !first {
				b.WriteByte(' ')
			}
			first = false

			b.Write(strconv.AppendUint(number[:0], uint64(y), 10))
			b.WriteString(":{C:")

			if board[x][y].Col {
				b.WriteByte('t')
			} else {
				b.WriteByte('f')
			}

			b.WriteString(" A:")
			b.Write(strconv.AppendUint(number[:0], uint64(board[x][y].Ang), 10))
			b.WriteByte('}')
		}
		b.WriteByte(']')
	}

	b.WriteString("] tK:{")
	b.WriteString(fmt.Sprintf("%v", s.trueKing)) // TODO replace with something optimized and shorter in bytes
	b.WriteString(" falseKing:")
	b.WriteString(fmt.Sprintf("%v", s.falseKing))
	b.WriteByte('}')
	return b.String()
}

func loadMemory(path string) (memory, error) {
	var loaded = make(memory)
	readFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	decoder := gob.NewDecoder(readFile)
	if err = decoder.Decode(&loaded); err != nil {
		return memory{}, err
	}
	if err := readFile.Close(); err != nil {
		return memory{}, err
	}
	return loaded, nil
}

func (data *memory) saveMemory(path string) error {
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
	superMap *memory, // Memory of past choice
	isOrigin bool, // true if this is the first time minimax was called.
	b *Board,
	depth uint8, // Search depth in tree
	alpha, beta int8, // used for alpha beta pruning
	who, isMaxingPlayer bool, // who is the color of the player
	maxKing, minKing KingPos,
) (int8, moveChoice, error) {

	// Check if this instance has already been computed
	strState := fmt.Sprint(state{*b, maxKing, minKing})
	currentConditions := conditions{depth, isMaxingPlayer}
	computedOutcome, exists := (*superMap)[strState][currentConditions]
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
		moves, err := b.GetAllMoves(maxKing)
		if err != nil {
			return 0, moveChoice{}, err
		}

		choice := moveChoice{}

		// Iterate over all possible piece moves
		for from, tos := range moves {
			for _, newMove := range tos {
				for angle := range uint8(6) {
					bClone := b.clone()
					// Excecute the move
					bClone[newMove.X][newMove.Y] = rotatePiece(bClone[from.X][from.Y], angle)
					bClone[from.X][from.Y] = Piece{}

					// Do scoring stuff
					eval, _, err := minimax(superMap, false, &bClone, depth-1, alpha, beta, !who, false, maxKing, minKing)
					if err != nil {
						return 0, moveChoice{}, err
					}
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
			eval, _, err := minimax(superMap, false, b, depth-1, alpha, beta, !who, false, KingPos{kMove, !maxKing.Ontop, maxKing.Col}, minKing)
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

		if _, exists = (*superMap)[strState]; !exists {
			(*superMap)[strState] = make(map[conditions]outcome)
		}
		(*superMap)[strState][currentConditions] = outcome{maxEval, choice}
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
		moves, err := b.GetAllMoves(minKing)
		if err != nil {
			return 0, moveChoice{}, err
		}

		choice := moveChoice{}

		// Iterate over all possible piece moves
		for from, tos := range moves {
			for _, newMove := range tos {
				for angle := range uint8(6) {
					bClone := b.clone()
					// Excecute the move
					bClone[newMove.X][newMove.Y] = rotatePiece(bClone[from.X][from.Y], angle)
					bClone[from.X][from.Y] = Piece{}

					// Do scoring stuff
					eval, _, err := minimax(superMap, false, &bClone, depth-1, alpha, beta, !who, true, maxKing, minKing)
					if err != nil {
						return 0, moveChoice{}, err
					}
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
			eval, _, err := minimax(superMap, false, b, depth-1, alpha, beta, !who, true, maxKing, KingPos{kMove, !minKing.Ontop, minKing.Col})
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

		if _, exists = (*superMap)[strState]; !exists {
			(*superMap)[strState] = make(map[conditions]outcome)
		}
		(*superMap)[strState][currentConditions] = outcome{minEval, choice}
		return minEval, choice, nil
	}
}
