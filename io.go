package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var stdinScanner *bufio.Scanner = bufio.NewScanner(os.Stdin)

func ReadLine() (string, error) {
	if stdinScanner.Scan() {
		return stdinScanner.Text(), nil
	}
	return "", stdinScanner.Err()
}

func AskForInputPosition(game *Game) (Position, error) {
	fmt.Print("Turn ", game.Step()/2+1, ` - Your turn(type "q" or "quit" to exit): `)
	pos := InvalidPosition
	for pos == InvalidPosition {
		input, err := ReadLine()
		if err != nil {
			return InvalidPosition, err
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		inputUpper := strings.ToUpper(input)
		if inputUpper == "Q" || inputUpper == "QUIT" {
			return InvalidPosition, nil
		}
		pos, err = ParsePosition(input)
		if err != nil {
			fmt.Println(err)
			fmt.Print(`Please input again(type "q" or "quit" to exit): `)
			pos = InvalidPosition
			continue
		}
		isLegal, hint, err := IsLegal(game.Settings.Rule, game.Step()+1, pos)
		if err != nil {
			return InvalidPosition, err
		}
		if !isLegal {
			fmt.Println("Position", pos, "is illegal.")
			if hint != "" {
				fmt.Println(hint)
			}
			fmt.Print(`Please input again(type "q" or "quit" to exit): `)
			pos = InvalidPosition
		}
	}
	return pos, nil
}

func PrintBoardToString(b map[Position]Piece, bpSettings *BoardPrintSettings) (
	string, error) {
	var ec, bc, wc string
	var sln bool
	if b == nil {
		b = make(map[Position]Piece) // An empty board.
	}
	if bpSettings != nil {
		ec = bpSettings.EmptyChar
		bc = bpSettings.BlackChar
		wc = bpSettings.WhiteChar
		sln = bpSettings.DoesShowLineNumber
	} else {
		ec = "."
		bc = "x"
		wc = "o"
		sln = true
	}
	numW := len(b) / 2
	numB := len(b) - numW
	capacity := (NumPosition-numB-numW)*len(ec) + numB*len(bc) + numW*len(wc) +
		NumPosition - 1 // Including '\n' and ' ' per line.
	if sln {
		capacity += BoardSize * 3 // Columns and space per row.
		// Rows:
		if BoardSize >= 10 && BoardSize < 100 {
			capacity += 9 + (BoardSize-9)*2
		} else if BoardSize < 10 {
			capacity += BoardSize
		} else {
			capacity += BoardSize * 3
		}
	}
	var builder strings.Builder
	builder.Grow(capacity)
	// fmt.Println("cap", builder.Cap())
	if sln {
		for i := 0; i < BoardSize; i++ {
			builder.WriteRune('A' + rune(i))
			if i < BoardSize-1 {
				builder.WriteRune(' ')
			} else {
				builder.WriteRune('\n')
			}
		}
	}
	for y := 0; y < BoardSize; y++ {
		for x := 0; x < BoardSize; x++ {
			p, err := GetPosition(x, y, false)
			if err != nil {
				return "", err
			}
			switch b[p] {
			case 0:
				builder.WriteString(ec)
			case Black:
				builder.WriteString(bc)
			case White:
				builder.WriteString(wc)
			default:
				return "", fmt.Errorf("unknown piece on board: %d", b[p])
			}
			if x < BoardSize-1 {
				builder.WriteRune(' ')
			}
		}
		if sln {
			builder.WriteRune(' ')
			builder.WriteString(strconv.Itoa(y + 1))
		}
		if y < BoardSize-1 {
			builder.WriteRune('\n')
		}
	}
	// fmt.Println("len", builder.Len())
	return builder.String(), nil
}

func PrintWelcome(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	bar := "------------------------------------------------------------"
	fmt.Fprintln(w, bar)
	fmt.Fprintln(w, "Welcome to Gomoku game!")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  You need input coordinates to place your stone.")
	fmt.Fprintln(w, "    Coordinates format: Letter(for column)+Number(for row)")
	fmt.Fprintln(w, `    e.g. "H8" is the center of the board`)
	fmt.Fprintln(w, "  You can change game settings in file:")
	fmt.Fprintln(w, "   ", SettingsPath)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "                           Developed by Yuan GAO.")
	fmt.Fprintln(w, bar)
}
