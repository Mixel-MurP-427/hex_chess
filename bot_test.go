package hexchess

import (
	"fmt"
	"math"
	"os"
	"testing"
)

func TestBot(t *testing.T) {

	fmt.Println(gobName)
	board := SetupBoard(false)

	tK := InitKing(true)

	fK := InitKing(false)

	theMemory, err := loadMemory(gobName)
	fmt.Println(err)
	// theMemory := memory{}

	result, botChoice, err := minimax(&theMemory, true, &board, 4, math.MinInt8, math.MaxInt8, true, true, tK, fK)

	fmt.Println(result)
	fmt.Println(botChoice)
	fmt.Println(err)
	fmt.Println(theMemory.saveMemory(gobName))
}

func TestReadGob(t *testing.T) {
	myObj, err := loadMemory(gobName)
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
