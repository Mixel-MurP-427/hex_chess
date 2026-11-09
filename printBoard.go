package hexchess

import (
	"fmt"
	"slices"
)

const (
	reset  = "\033[0m"
	blue   = "\033[0;36m"
	orange = "\033[93m"
	gray   = "\033[38;5;238m"
)

func (b *Board) printBoard() {
	pb := [9][9]uint8{{0}}

	xs := [9]uint8{4, 3, 2, 1, 0, 0, 0, 0, 0}
	ys := [9]uint8{8, 8, 8, 8, 8, 7, 6, 5, 4}

	for i := range b {
		spaces, err := b.getSpacesInRow(Coord{xs[i], ys[i]}, 0)
		if err != nil {
			fmt.Println(err)
		}
		spaces = slices.Concat([]Coord{{xs[i], ys[i]}}, spaces)

		for j, space := range spaces {
			p := b[space.x][space.y]

			if len(p.ang) == 0 {
				pb[i][j] = 3
				continue
			}
			if p.col {
				pb[i][j] = 1
			} else {
				pb[i][j] = 2
			}
		}
	}

	pre := [9]string{
		"    ",
		"   ",
		"  ",
		" ",
		"",
		" ",
		"  ",
		"   ",
		"    ",
	}

	post := [9]string{
		"3",
		"2",
		"1",
		"0",
		"",
		"8",
		"7",
		"6",
		"5",
	}

	fmt.Println(gray + "     8 7 6 5 4")

	for i, row := range pb {
		line := pre[i]
	chLoop:
		for _, char := range row {
			switch char {
			case 0:
				break chLoop
			case 1:
				line += blue + "X "
			case 2:
				line += orange + "X "
			case 3:
				line += reset + "+ "
			default:
				fmt.Println("Error!")
			}
		}
		line += gray + post[i]
		fmt.Println(line)
	}
	fmt.Println("     0 1 2 3 4" + reset)

}
