package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/donyori/gorecover"
)

func main() {
	err := gorecover.Recover(func() {
		err := body()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
}

// For debug.
/*
func main() {
	err := body()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
}*/

func body() error {
	rand.Seed(time.Now().UnixNano())

	settings, err := LoadSettings()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if settings == nil {
		settings = NewSettings()
		err = StoreSettings(settings)
		if err != nil {
			// Just warning but not exit.
			fmt.Fprintln(os.Stderr, "Try to store settings to", SettingsPath,
				"but failed. Error:", err)
		}
	}

	game, err := NewGame(settings)
	if err != nil {
		return err
	}
	defer game.TearDown()

	boardStr, err := PrintBoardToString(nil, settings.Io.BoardPrint)
	if err != nil {
		return err
	}
	PrintWelcome(nil)
	fmt.Println()
	fmt.Println(boardStr)
	fmt.Println()

	var pos Position
	player := Black
	for !game.IsTerminal() {
		if player&settings.Ai.AiPiece > 0 {
			fmt.Print("Turn ", game.Step()/2+1, " - AI's turn: ")
			pos, err = game.PlaceByAi()
			if err != nil {
				return err
			}
			fmt.Println(pos)
		} else {
			// Ask for user input.
			pos, err = AskForInputPosition(game)
			if err != nil {
				return err
			}
			if pos == InvalidPosition {
				// User want to quit the game.
				return nil
			}
			err = game.PlaceByUser(pos)
			if err != nil {
				return err
			}
		}
		boardStr, err = PrintBoardToString(game.Board, settings.Io.BoardPrint)
		if err != nil {
			return err
		}
		fmt.Println()
		fmt.Println(boardStr)
		fmt.Println()
		if player == Black {
			player = White
		} else {
			player = Black
		}
	}
	fmt.Println("Game over. Winner:", game.Outcome)
	return nil
}
