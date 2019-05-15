package main

import "testing"

func TestPrintBoardToString(t *testing.T) {
	b := make(map[Position]Piece)
	mid, err := ParsePosition("H8")
	if err != nil {
		t.Fatal(err)
	}
	b[mid] = Black
	p, err := mid.Move(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	b[p] = White
	p, err = mid.Move(-1, 1)
	if err != nil {
		t.Fatal(err)
	}
	b[p] = Black
	s, err := PrintBoardToString(b, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("\n" + s)
	t.Log("len", len(s))
}
