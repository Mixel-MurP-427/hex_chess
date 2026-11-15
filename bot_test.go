package hexchess

import (
	"fmt"
	"os"
	"testing"
)

func TestBot(t *testing.T) {

	fmt.Println(gobName)
	board := SetupBoard(false)

	tK := InitKing(true)

	fK := InitKing(false)

	theMemory, err := loadBotChoices(gobName)
	fmt.Println(err)
	// theMemory := memory{}

	//result, botChoice, err := minimax(&theMemory, true, &board, 3, math.MinInt8, math.MaxInt8, true, true, tK, fK)
	choice, err := BotChoose(4, &board, tK, fK, &theMemory)

	fmt.Println(choice)
	fmt.Println(err)
	fmt.Println(theMemory.saveBotChoices(gobName))
}

func TestReadGob(t *testing.T) {
	myObj, err := loadBotChoices(gobName)
	fmt.Println(err)

	// Convert struct to formatted JSON bytes
	data := fmt.Sprintln(myObj)
	if err != nil {
		panic(err)
	}

	// Write bytes to a text file
	err = os.WriteFile("thegob.txt", []byte(data), 0644)
	if err != nil {
		panic(err)
	}
}

func TestSprint(t *testing.T) {
	board := SetupBoard(false)
	fmt.Println(&board)
	fmt.Println(KingPos{Coord{6, 7}, true, true})
	fmt.Println(state{KingPos{}, KingPos{}, &board})
}
