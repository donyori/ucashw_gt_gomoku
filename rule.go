package main

import (
	"errors"
	"strings"
)

type Rule int8

const (
	StandardGomoku Rule = iota + 1
	GomokuPro
)

var ruleStrings = [...]string{
	"Unknown",
	"StandardGomoku",
	"Gomoku-Pro",
}

func ParseRule(s string) Rule {
	for i := range ruleStrings {
		if strings.EqualFold(s, ruleStrings[i]) {
			return Rule(i)
		}
	}
	return 0 // Stands for "Unknown".
}

func (r Rule) String() string {
	if r < StandardGomoku || r > GomokuPro {
		return ruleStrings[0]
	}
	return ruleStrings[r]
}

func (r Rule) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *Rule) UnmarshalText(text []byte) error {
	*r = ParseRule(string(text))
	return nil
}

func IsLegal(rule Rule, step uint, pos Position) (
	isLegal bool, hint string, err error) {
	switch rule {
	case StandardGomoku:
		return isLegalStdGomoku(step, pos)
	case GomokuPro:
		return isLegalGomokuPro(step, pos)
	default:
		return false, "", ErrUnknownRule
	}
}

func isLegalStdGomoku(step uint, pos Position) (
	isLegal bool, hint string, err error) {
	if step == 0 {
		panic(errors.New("step is zero"))
	}
	if pos.IsOutOfRange() {
		return false, "Position is outside the board.", nil
	}
	return true, "", nil
}

func isLegalGomokuPro(step uint, pos Position) (
	isLegal bool, hint string, err error) {
	if step != 1 && step != 3 {
		return isLegalStdGomoku(step, pos)
	}
	x, y := pos.XOffset(), pos.YOffset()
	if step == 1 {
		if x != 0 || y != 0 {
			return false, "First step must be at H8.", nil
		}
		return true, "", nil
	} else { // step == 3
		ia, h, e := isLegalStdGomoku(step, pos)
		if !ia {
			return ia, h, e
		}
		if x >= -2 && x <= 2 && y >= -2 && y <= 2 {
			return false, "Third step must be outside the central 5×5 area.", nil
		}
		return true, "", nil
	}
}
