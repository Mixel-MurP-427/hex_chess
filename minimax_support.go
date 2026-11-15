package hexchess

import (
	"encoding/gob"
	"fmt"
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
