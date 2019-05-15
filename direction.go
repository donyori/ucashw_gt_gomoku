package main

type Direction int8

const (
	Left Direction = iota + 1
	LeftUp
	Up
	RightUp
	Right
	RightDown
	Down
	LeftDown
)
