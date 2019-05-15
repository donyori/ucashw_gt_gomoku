package main

import (
	"errors"
	"fmt"
)

type UnknownPositionError struct {
	s string
}

type PositionOutOfRangeError struct {
	x, y int
}

var ErrUnknownRule error = errors.New("rule is unknown")

func NewUnknownPositionError(s string) error {
	return &UnknownPositionError{s: s}
}

func (upe *UnknownPositionError) Error() string {
	return fmt.Sprintf("position %q is unknown", upe.s)
}

func NewPositionOutOfRangeError(x, y int) error {
	if x >= 0 && x < BoardSize && y >= 0 && y < BoardSize {
		panic(fmt.Errorf("position(x: %d, y: %d) is NOT out of range(0-%d), "+
			"but treat it as an error", x, y, BoardSize-1))
	}
	return &PositionOutOfRangeError{x: x, y: y}
}

func (pore *PositionOutOfRangeError) Error() string {
	return fmt.Sprintf("position is out of range(0-%d), x: %d, y: %d",
		BoardSize-1, pore.x, pore.y)
}
