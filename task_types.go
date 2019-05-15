package main

type GetValidPosTask struct {
	LookupPieceFn func(pos Position) Piece
	Step          uint
	Pos           Position
	Output        chan<- Position
}

type NodeAndUct struct {
	Node *MonteCarloTreeNode
	Uct  float64
}

type CalcUctTask struct {
	Node   *MonteCarloTreeNode
	Output chan<- *NodeAndUct
}

type CkOutcomeTask struct {
	LookupPieceFn func(pos Position) Piece
	Pos           Position
	Dir           Direction
	CntrAddr      *uint32
}

type WaitAndCloseTask struct {
	WaitTgt  interface{ Wait() }
	CloseTgt interface{}
}
