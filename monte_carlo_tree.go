package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/donyori/goctpf"
)

type MonteCarloTreeNode struct {
	Game *Game

	Parent, PrevSibling, LastChild *MonteCarloTreeNode

	Step uint
	Pos  Position

	NumWin uint64
	NumSim uint64

	unexpPos []Position
}

func NewMonteCarloTree(game *Game, step uint, pos Position) (
	*MonteCarloTreeNode, error) {
	if game == nil {
		panic(errors.New("game is nil"))
	}
	node := &MonteCarloTreeNode{
		Game: game,
		Step: step,
		Pos:  pos,
	}
	if step > 0 {
		piece := game.CheckOutcome(node.LookupPiece, pos)
		switch piece {
		case 0, Both:
			node.unexpPos = game.GetValidPositionsAsSlice(
				node.LookupPiece, step+1, true, true)
		case Black, White:
			node.unexpPos = nil
		default:
			return nil, fmt.Errorf("cannot check outcome on position %v", pos)
		}
	} else {
		rule := game.Settings.Rule
		isLegal, _, err := IsLegal(rule, 1, CenterPosition)
		if err != nil {
			return nil, err
		}
		if isLegal {
			// If "H8" is legal, place here.
			node.unexpPos = []Position{CenterPosition}
		} else {
			node.unexpPos = make([]Position, 0, NumPosition)
			for p := MinPosition; p <= MaxPosition; p++ {
				isLegal, _, err = IsLegal(rule, 1, p)
				if err != nil {
					return nil, err
				}
				if isLegal {
					node.unexpPos = append(node.unexpPos, p)
				}
			}
			if len(node.unexpPos) != cap(node.unexpPos) {
				node.unexpPos = node.unexpPos[:len(node.unexpPos):len(node.unexpPos)]
			}
			rand.Shuffle(len(node.unexpPos), func(i, j int) {
				node.unexpPos[i], node.unexpPos[j] = node.unexpPos[j], node.unexpPos[i]
			})
		}
	}
	return node, nil
}

func (mctn *MonteCarloTreeNode) LookupPiece(pos Position) Piece {
	if mctn == nil || pos.IsOutOfRange() {
		return InvalidPiece
	}
	piece := mctn.Game.LookupPiece(pos)
	if piece != 0 {
		return piece
	}
	for node := mctn; node != nil && node.Step > 0; node = node.Parent {
		if node.Pos == pos {
			if node.Step%2 == 1 {
				return Black
			} else {
				return White
			}
		}
	}
	return 0 // Stands for "None".
}

func (mctn *MonteCarloTreeNode) IsTerminal() bool {
	return mctn == nil || (len(mctn.unexpPos) == 0 && mctn.LastChild == nil)
}

func (mctn *MonteCarloTreeNode) GetBestNumSimChild() *MonteCarloTreeNode {
	if mctn == nil || mctn.LastChild == nil {
		return nil
	}
	best := mctn.LastChild
	var n float64 = 1.
	for node := best.PrevSibling; node != nil; node = node.PrevSibling {
		if node.NumSim > best.NumSim {
			n = 1.
			best = node
		} else if node.NumSim == best.NumSim {
			// Pick one of the best NumSim children randomly, with equal probability.
			n++
			if rand.Float64() < 1./n {
				best = node
			}
		}
	}
	return best
}

// Upper Confidence Bound 1 applied to trees.
func (mctn *MonteCarloTreeNode) Uct() float64 {
	if mctn == nil {
		return 0.
	}
	if mctn.NumSim == 0 || mctn.Parent == nil {
		return math.Inf(1)
	}
	w := float64(mctn.NumWin)
	n := float64(mctn.NumSim)
	nParent := float64(mctn.Parent.NumSim)
	return w/n + mctn.Game.Settings.Ai.UctParamC*math.Sqrt(math.Log(nParent)/n)
}

func (mctn *MonteCarloTreeNode) GetBestUctChild() *MonteCarloTreeNode {
	if mctn == nil || mctn.LastChild == nil {
		return nil
	}
	tg := goctpf.NewTaskGroup(nil, nil)
	outputChan := make(chan *NodeAndUct, NumPosition-len(mctn.unexpPos))
	for node := mctn.LastChild; node != nil; node = node.PrevSibling {
		mctn.Game.SubmitCalcUctTask(tg.WrapTask(&CalcUctTask{
			Node:   node,
			Output: outputChan,
		}))
	}
	mctn.Game.SubmitWaitAndCloseTask(&WaitAndCloseTask{
		WaitTgt:  tg,
		CloseTgt: outputChan,
	})
	cmpThold := mctn.Game.Settings.Ai.UctCmpThold
	if cmpThold == 0. {
		cmpThold = Epsilon
	}
	best := &NodeAndUct{Uct: -math.MaxFloat64}
	var n float64
	// For debug:
	//fmt.Println("Waiting for output")
	for output := range outputChan {
		// For debug:
		//fmt.Printf("best %#v\n", best)
		if output.Uct > best.Uct+cmpThold {
			// For debug:
			//fmt.Printf("Goin case 1, output: %#v\n", output)
			n = 1.
			best = output
		} else if output.Uct > best.Uct-cmpThold {
			// For debug:
			//fmt.Printf("Goin case 2, output: %#v\n", output)
			// Pick one of the best UCT children randomly, with equal probability.
			n++
			if rand.Float64() < 1./n {
				// For debug:
				//fmt.Println("Update in case 2.")
				best = output
			}
		}
	}
	return best.Node
}

func (mctn *MonteCarloTreeNode) IsFullyExpanded() bool {
	return mctn == nil || len(mctn.unexpPos) == 0
}

func (mctn *MonteCarloTreeNode) Expand() (*MonteCarloTreeNode, error) {
	if mctn.IsFullyExpanded() {
		return nil, nil
	}
	last := len(mctn.unexpPos) - 1
	pos := mctn.unexpPos[last]

	node := &MonteCarloTreeNode{
		Game:        mctn.Game,
		Parent:      mctn,
		PrevSibling: mctn.LastChild,
		Step:        mctn.Step + 1,
		Pos:         pos,
	}
	piece := mctn.Game.CheckOutcome(node.LookupPiece, pos)
	switch piece {
	case 0, Both:
		node.unexpPos = mctn.Game.GetValidPositionsAsSlice(
			node.LookupPiece, node.Step+1, true, true)
	case Black, White:
		node.unexpPos = nil
	default:
		return nil, fmt.Errorf("cannot check outcome on position %v", pos)
	}

	mctn.LastChild = node
	mctn.unexpPos[last] = InvalidPosition
	if last > 0 {
		mctn.unexpPos = mctn.unexpPos[:last]
	} else {
		mctn.unexpPos = nil
	}
	return node, nil
}

func (mctn *MonteCarloTreeNode) Rollout() Piece {
	if mctn == nil {
		return InvalidPiece
	}
	exBoard := make(map[Position]Piece)
	isBlack := mctn.Step%2 == 1
	for node := mctn; node != nil && node.Step > 0; node = node.Parent {
		if isBlack {
			exBoard[node.Pos] = Black
		} else {
			exBoard[node.Pos] = White
		}
		isBlack = !isBlack
	}
	lookupPieceFn := func(pos Position) Piece {
		if pos.IsOutOfRange() {
			return InvalidPiece
		}
		piece := mctn.Game.LookupPiece(pos)
		if piece != 0 {
			return piece
		}
		return exBoard[pos]
	}
	if mctn.IsTerminal() {
		return mctn.Game.CheckOutcome(lookupPieceFn, mctn.Pos)
	}
	var n float64
	step := mctn.Step
	isBlack = step%2 == 1
	pos := mctn.Pos
	var outcome Piece
	for outcome == 0 {
		n = 0.
		step++
		isBlack = !isBlack
		vpc := mctn.Game.GetValidPositions(lookupPieceFn, step)
		for vp := range vpc {
			n++
			if rand.Float64() <= 1./n {
				// Pick one of the valid position randomly, with equal probability.
				pos = vp
			}
		}
		if n < Epsilon {
			// n == 0.
			// i.e. Outcome is draw.
			return 0
		}
		if isBlack {
			exBoard[pos] = Black
		} else {
			exBoard[pos] = White
		}
		outcome = mctn.Game.CheckOutcome(lookupPieceFn, pos)
	}
	return outcome
}

func (mctn *MonteCarloTreeNode) BackPropagate(outcome Piece) error {
	var isWin bool
	switch outcome {
	case 0, Both:
		outcome = 0
	case Black:
		isWin = mctn.Step%2 == 1
	case White:
		isWin = mctn.Step%2 == 0
	default:
		return fmt.Errorf("outcome(%b) is invalid", outcome)
	}
	for node := mctn; node != nil; node = node.Parent {
		if isWin {
			node.NumWin++
		}
		node.NumSim++
		if outcome != 0 {
			isWin = !isWin
		}
	}
	return nil
}

func (mctn *MonteCarloTreeNode) TakeOut() {
	if mctn == nil || mctn.Parent == nil {
		// Is nil or already as root, just return.
		return
	}
	parent := mctn.Parent
	sibling := mctn.PrevSibling
	mctn.Parent = nil
	mctn.PrevSibling = nil
	child := parent.LastChild
	if child == mctn {
		parent.LastChild = sibling
		return
	}
	s := child.PrevSibling
	for s != mctn {
		child = s
		s = child.PrevSibling
	}
	if child == nil {
		return
	}
	child.PrevSibling = sibling
}

// Selection and expansion steps of Monte Carlo tree search.
func (mctn *MonteCarloTreeNode) Traverse() (*MonteCarloTreeNode, error) {
	if mctn == nil {
		return nil, nil
	}
	node := mctn
	for node.IsFullyExpanded() && !node.IsTerminal() {
		node = node.GetBestUctChild()
	}
	if node.IsTerminal() {
		return node, nil
	}
	return node.Expand()
}

// Perform one simulation(including selection, expansion, rollout and backpropagation)
//   of Monte Carlo tree search.
// Return the elapsed time and occured error.
func (mctn *MonteCarloTreeNode) Simulate() (
	elapsedTime time.Duration, err error) {
	if mctn == nil {
		return
	}
	startTime := time.Now()
	defer func() {
		elapsedTime = time.Since(startTime)
	}()
	var node *MonteCarloTreeNode
	node, err = mctn.Traverse()
	if err != nil {
		return
	}
	outcome := node.Rollout()
	err = node.BackPropagate(outcome)
	return
}

func (mctn *MonteCarloTreeNode) MonteCarloTreeSearch() (
	bestChild *MonteCarloTreeNode, err error) {
	if mctn == nil || mctn.IsTerminal() {
		return mctn, nil
	}
	startTime := time.Now()
	var numSim float64
	var halfAvgElapsedTime float64
	for float64(mctn.Game.Settings.Ai.MctsTimeLimit-time.Since(startTime)) >
		halfAvgElapsedTime {
		elapsedTime, err := mctn.Simulate()
		if err != nil {
			return nil, err
		}
		numSim++
		halfAvgElapsedTime = (halfAvgElapsedTime*(numSim-1.) +
			float64(elapsedTime)/2.) / numSim
	}
	return mctn.GetBestNumSimChild(), nil
}
