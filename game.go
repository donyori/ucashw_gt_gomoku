package main

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sync/atomic"

	"github.com/donyori/goctpf"
	"github.com/donyori/goctpf/idtpf/dfw"
	"github.com/donyori/goctpf/prefab"
)

type Game struct {
	Settings *Settings

	History []Position
	Board   map[Position]Piece
	Outcome Piece

	mctRoot *MonteCarloTreeNode

	waitAndCloseInputChan chan<- interface{}
	waitAndCloseDoneChan  <-chan struct{}
	getValidPosInputChan  chan<- interface{}
	getValidPosDoneChan   <-chan struct{}
	calcUctInputChan      chan<- interface{}
	calcUctDoneChan       <-chan struct{}
	ckOutcomeInputChan    chan<- interface{}
	ckOutcomeDoneChan     <-chan struct{}
}

func NewGame(settings *Settings) (*Game, error) {
	if settings == nil {
		settings = NewSettings()
	}
	g := &Game{Settings: settings}
	root, err := NewMonteCarloTree(g, 0, InvalidPosition)
	if err != nil {
		return nil, err
	}
	wacic := make(chan interface{}, 1)
	gvpic := make(chan interface{}, 1)
	cuic := make(chan interface{}, 1)
	coic := make(chan interface{}, 1)
	g.History = make([]Position, 0, NumPosition)
	g.Board = make(map[Position]Piece)
	g.mctRoot = root
	g.waitAndCloseInputChan = wacic
	g.waitAndCloseDoneChan = dfw.StartEx(prefab.QueueTaskManagerMaker,
		g.waitAndCloseHandler, nil, nil, goctpf.WorkerSettings{Number: 3},
		wacic, nil)
	g.getValidPosInputChan = gvpic
	g.getValidPosDoneChan = dfw.StartEx(prefab.StackTaskManagerMaker,
		g.getValidPosHandler, nil, nil, *settings.Worker, gvpic, nil)
	g.calcUctInputChan = cuic
	g.calcUctDoneChan = dfw.StartEx(prefab.StackTaskManagerMaker,
		g.calcUctHandler, nil, nil, *settings.Worker, cuic, nil)
	g.ckOutcomeInputChan = coic
	g.ckOutcomeDoneChan = dfw.StartEx(prefab.StackTaskManagerMaker,
		g.ckOutcomeHandler, nil, nil, *settings.Worker, coic, nil)
	return g, nil
}

func (g *Game) IsTearDown() bool {
	return g == nil || g.mctRoot == nil
}

func (g *Game) TearDown() {
	if g.getValidPosInputChan != nil {
		close(g.getValidPosInputChan)
		g.getValidPosInputChan = nil
	}
	if g.calcUctInputChan != nil {
		close(g.calcUctInputChan)
		g.calcUctInputChan = nil
	}
	if g.ckOutcomeInputChan != nil {
		close(g.ckOutcomeInputChan)
		g.ckOutcomeInputChan = nil
	}
	if g.waitAndCloseInputChan != nil {
		close(g.waitAndCloseInputChan)
		g.waitAndCloseInputChan = nil
	}
	if g.getValidPosDoneChan != nil {
		<-g.getValidPosDoneChan
		g.getValidPosDoneChan = nil
	}
	if g.calcUctDoneChan != nil {
		<-g.calcUctDoneChan
		g.calcUctDoneChan = nil
	}
	if g.ckOutcomeDoneChan != nil {
		<-g.ckOutcomeDoneChan
		g.ckOutcomeDoneChan = nil
	}
	if g.waitAndCloseDoneChan != nil {
		<-g.waitAndCloseDoneChan
		g.waitAndCloseDoneChan = nil
	}

	g.mctRoot = nil
}

func (g *Game) IsTerminal() bool {
	return g.IsTearDown() || g.Outcome != 0 || g.mctRoot.IsTerminal()
}

func (g *Game) Step() uint {
	return uint(len(g.History))
}

func (g *Game) NextTurn() Piece {
	if g.IsTerminal() {
		return InvalidPiece
	}
	step := g.Step()
	// Step is for current, return value is for next.
	if step%2 == 0 {
		return Black
	} else {
		return White
	}
}

func (g *Game) PlaceByUser(pos Position) error {
	if g.IsTearDown() {
		panic(errors.New("game is already tear-down"))
	}
	if g.IsTerminal() {
		panic(errors.New("game is terminal"))
	}
	if pos.IsOutOfRange() {
		panic(errors.New("position is out of range"))
	}
	g.updateHistoryAndBoard(pos)
	for node := g.mctRoot.LastChild; node != nil; node = node.PrevSibling {
		if node.Pos == pos {
			g.mctRoot = node
			node.TakeOut()
			if node.IsTerminal() {
				g.Outcome = g.CheckOutcome(nil, pos)
			}
			return nil
		}
	}
	// The case: pos is NOT valid but legal, or root is not fully expanded!
	step := g.mctRoot.Step + 1
	root, err := NewMonteCarloTree(g, step, pos)
	if err != nil {
		return err
	}
	g.mctRoot = root
	if root.IsTerminal() {
		g.Outcome = g.CheckOutcome(nil, pos)
	}
	return nil
}

func (g *Game) PlaceByAi() (Position, error) {
	if g.IsTearDown() {
		panic(errors.New("game is already tear-down"))
	}
	if g.IsTerminal() {
		panic(errors.New("game is terminal"))
	}
	if g.NextTurn()&g.Settings.Ai.AiPiece == 0 {
		panic(errors.New("it's not AI's turn"))
	}
	best, err := g.mctRoot.MonteCarloTreeSearch()
	if err != nil {
		return InvalidPosition, err
	}
	if best == nil {
		return InvalidPosition, errors.New(
			"cannot find a position to place stone")
	}
	g.updateHistoryAndBoard(best.Pos)
	g.mctRoot = best
	best.TakeOut()
	if best.IsTerminal() {
		g.Outcome = g.CheckOutcome(nil, best.Pos)
	}
	return best.Pos, nil
}

func (g *Game) LookupPiece(pos Position) Piece {
	if g == nil || pos.IsOutOfRange() {
		return InvalidPiece
	}
	return g.Board[pos]
}

func (g *Game) SubmitWaitAndCloseTask(task *WaitAndCloseTask) {
	if g.IsTearDown() {
		panic(errors.New("game is already tear-down"))
	}
	if task == nil {
		panic(errors.New("task is nil"))
	}
	if task.WaitTgt == nil {
		panic(errors.New("WaitAndCloseTask.WaitTgt is nil"))
	}
	if task.CloseTgt == nil {
		panic(errors.New("WaitAndCloseTask.CloseTgt is nil"))
	}
	t := reflect.TypeOf(task.CloseTgt)
	if t.Kind() != reflect.Chan || t.ChanDir()&reflect.SendDir == 0 {
		panic(errors.New("WaitAndCloseTask.CloseTgt is not send chan"))
	}
	g.waitAndCloseInputChan <- task
}

func (g *Game) GetValidPositions(lookupPieceFn func(pos Position) Piece,
	step uint) <-chan Position {
	if g.IsTearDown() {
		panic(errors.New("game is already tear-down"))
	}
	if step == 0 {
		panic(errors.New("step is zero"))
	}
	if lookupPieceFn == nil {
		lookupPieceFn = g.LookupPiece
	}

	outputChan := make(chan Position, NumPosition)
	tg := goctpf.NewTaskGroup(nil, nil)
	for p := MinPosition; p <= MaxPosition; p++ {
		g.getValidPosInputChan <- tg.WrapTask(&GetValidPosTask{
			LookupPieceFn: lookupPieceFn,
			Step:          step,
			Pos:           p,
			Output:        outputChan,
		})
	}
	g.SubmitWaitAndCloseTask(&WaitAndCloseTask{
		WaitTgt:  tg,
		CloseTgt: outputChan,
	})
	return outputChan
}

func (g *Game) GetValidPositionsAsSlice(lookupPieceFn func(pos Position) Piece,
	step uint, doesShrink, doesShuffle bool) []Position {
	var vps []Position
	vpc := g.GetValidPositions(lookupPieceFn, step)
	for vp := range vpc {
		vps = append(vps, vp)
	}
	if doesShrink && len(vps) != cap(vps) {
		// Shrink the array:
		vps = vps[:len(vps):len(vps)]
	}
	if doesShuffle {
		rand.Shuffle(len(vps), func(i int, j int) {
			vps[i], vps[j] = vps[j], vps[i]
		})
	}
	return vps
}

func (g *Game) SubmitCalcUctTask(task interface{}) {
	if g.IsTearDown() {
		panic(errors.New("game is already tear-down"))
	}
	if task == nil {
		panic(errors.New("task is nil"))
	}
	var calcUctTask *CalcUctTask
	var ok bool
	switch task.(type) {
	case *goctpf.TaskGroupMember:
		t := task.(*goctpf.TaskGroupMember).Task
		calcUctTask, ok = t.(*CalcUctTask)
	case *CalcUctTask:
		calcUctTask = task.(*CalcUctTask)
		ok = true
	}
	if !ok {
		panic(fmt.Errorf(
			"task is neither a CalcUctTask nor a TaskGroupMember with a CalcUctTask, task type: %T",
			task))
	}
	if calcUctTask == nil {
		panic(errors.New("CalcUctTask is nil"))
	}
	if calcUctTask.Node == nil {
		panic(errors.New("CalcUctTask.Node is nil"))
	}
	if calcUctTask.Output == nil {
		panic(errors.New("CalcUctTask.Output is nil"))
	}
	g.calcUctInputChan <- task
}

func (g *Game) CheckOutcome(lookupPieceFn func(pos Position) Piece,
	pos Position) Piece {
	if g.IsTearDown() {
		panic(errors.New("game is already tear-down"))
	}
	if pos.IsOutOfRange() {
		return InvalidPiece
	}
	if lookupPieceFn == nil {
		lookupPieceFn = g.LookupPiece
	}
	piece := lookupPieceFn(pos)
	switch piece {
	case 0, Both:
		return 0
	case Black, White:
		// Do nothing here.
	default:
		return InvalidPiece
	}

	var hCntr, vCntr, dCntrLR, dCntrRL uint32
	tg := goctpf.NewTaskGroup(nil, nil)
	var cases = [...]struct {
		Dir  Direction
		Addr *uint32
	}{
		{Left, &hCntr},
		{LeftUp, &dCntrLR},
		{Up, &vCntr},
		{RightUp, &dCntrRL},
		{Right, &hCntr},
		{RightDown, &dCntrLR},
		{Down, &vCntr},
		{LeftDown, &dCntrRL},
	}
	for i := range cases {
		g.ckOutcomeInputChan <- tg.WrapTask(&CkOutcomeTask{
			LookupPieceFn: lookupPieceFn,
			Pos:           pos,
			Dir:           cases[i].Dir,
			CntrAddr:      cases[i].Addr,
		})
	}
	tg.Wait()

	if hCntr >= 4 || vCntr >= 4 || dCntrLR >= 4 || dCntrRL >= 4 {
		return piece
	}
	return 0
}

func (g *Game) updateHistoryAndBoard(pos Position) {
	g.History = append(g.History, pos)
	step := g.mctRoot.Step + 1
	if step%2 == 1 {
		g.Board[pos] = Black
	} else {
		g.Board[pos] = White
	}
}

func (g *Game) waitAndCloseHandler(workerNo int, task interface{},
	errBuf *[]error) (newTasks []interface{}, doesExit bool) {
	// Always return nil, false. So just use "return".
	t := task.(*WaitAndCloseTask)
	t.WaitTgt.Wait()
	reflect.ValueOf(t.CloseTgt).Close()
	return
}

func (g *Game) getValidPosHandler(workerNo int, task interface{},
	errBuf *[]error) (newTasks []interface{}, doesExit bool) {
	// Always return nil, false. So just use "return".
	distThold := g.Settings.Ai.ValidDistThold
	t := task.(*goctpf.TaskGroupMember).Task.(*GetValidPosTask)
	if t.LookupPieceFn(t.Pos) != 0 {
		// For debug:
		//fmt.Printf("Pos: %v, Step: %d - Invalid at A.\n", t.Pos, t.Step)
		return
	}
	if t.Step == 3 && g.Settings.Rule == GomokuPro && distThold < 2 {
		// Special case.
		distThold = 2
	}
	isLegal, _, err := IsLegal(g.Settings.Rule, t.Step, t.Pos)
	if !isLegal || err != nil {
		// For debug:
		//fmt.Printf("Pos: %v, Step: %d - Invalid at B.\n", t.Pos, t.Step)
		return
	}

	if t.Step == 1 || distThold == 0 {
		// First step can be at any legal position.
		// Treat distThold == 0 as no additional limit.
		if t.Pos != CenterPosition {
			// If "H8" is legal, only place at "H8".
			isLegal, _, err = IsLegal(g.Settings.Rule, 1, CenterPosition)
			if isLegal && err == nil {
				return
			}
		}
		t.Output <- t.Pos
		// For debug:
		//fmt.Printf("Pos: %v, Step: %d - Valid at C.\n", t.Pos, t.Step)
		return
	}

	x, y := t.Pos.X(), t.Pos.Y()
	left, right := x-int(distThold), x+int(distThold)
	top, bottom := y-int(distThold), y+int(distThold)
	if left < 0 {
		left = 0
	}
	if top < 0 {
		top = 0
	}
	if right >= BoardSize {
		right = BoardSize - 1
	}
	if bottom >= BoardSize {
		bottom = BoardSize - 1
	}
	posEndOffset := Position(right - left)
	var isValid bool
	for y = top; !isValid && y <= bottom; y++ {
		pos, err := GetPosition(left, y, false)
		if err != nil {
			continue
		}
		for posEnd := pos + posEndOffset; !isValid && pos <= posEnd; pos++ {
			if pos == t.Pos {
				continue
			}
			piece := t.LookupPieceFn(pos)
			if piece == Black || piece == White {
				isValid = true
			}
		}
	}

	if isValid {
		// For debug:
		//fmt.Printf("Pos: %v, Step: %d - Valid at D.\n", t.Pos, t.Step)
		t.Output <- t.Pos
	} /* else {
		// For debug:
		fmt.Printf("Pos: %v, Step: %d - Invalid at E.\n", t.Pos, t.Step)
	}*/
	return
}

func (g *Game) calcUctHandler(workerNo int, task interface{},
	errBuf *[]error) (newTasks []interface{}, doesExit bool) {
	// Always return nil, false. So just use "return".
	t := task.(*goctpf.TaskGroupMember).Task.(*CalcUctTask)
	uct := t.Node.Uct()
	t.Output <- &NodeAndUct{Node: t.Node, Uct: uct}
	return
}

func (g *Game) ckOutcomeHandler(workerNo int, task interface{},
	errBuf *[]error) (newTasks []interface{}, doesExit bool) {
	// Always return nil, false. So just use "return".
	t := task.(*goctpf.TaskGroupMember).Task.(*CkOutcomeTask)
	if atomic.LoadUint32(t.CntrAddr) >= 4 {
		// Already get the outcome, just return.
		return
	}
	pos := t.Pos
	piece := t.LookupPieceFn(pos)
	var cntr uint32
	i, j := pos.X(), pos.Y()
	var iDelta, jDelta, posDelta int
	posInt := int(pos)
	switch t.Dir {
	case Left:
		i--
		iDelta = -1
		posDelta = -1
	case LeftUp:
		i--
		j--
		iDelta = -1
		jDelta = -1
		posDelta = -1 - BoardSize
	case Up:
		j--
		jDelta = -1
		posDelta = -BoardSize
	case RightUp:
		i++
		j--
		iDelta = 1
		jDelta = -1
		posDelta = 1 - BoardSize
	case Right:
		i++
		iDelta = 1
		posDelta = 1
	case RightDown:
		i++
		j++
		iDelta = 1
		jDelta = 1
		posDelta = 1 + BoardSize
	case Down:
		j++
		jDelta = 1
		posDelta = BoardSize
	case LeftDown:
		i--
		j++
		iDelta = -1
		jDelta = 1
		posDelta = BoardSize - 1
	default:
		return
	}
	for i >= 0 && i < BoardSize && j >= 0 && j < BoardSize {
		posInt += posDelta
		pos = Position(posInt)
		if t.LookupPieceFn(pos) != piece {
			break
		}
		cntr++
		i += iDelta
		j += jDelta
	}
	if cntr > 0 {
		atomic.AddUint32(t.CntrAddr, cntr)
	}
	return
}
