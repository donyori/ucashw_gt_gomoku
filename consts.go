package main

import "math"

const BoardSize int = 15
const PositionOffset int = BoardSize / 2
const NumPosition int = BoardSize * BoardSize

// Position ∈ [1, 225]. 0 for invalid.
const (
	InvalidPosition Position = 0
	MinPosition     Position = 1
	MaxPosition              = Position(NumPosition)
	CenterPosition           = (MinPosition + MaxPosition) / 2
)

var Epsilon float64 = math.Nextafter(1., 2.) - 1.
