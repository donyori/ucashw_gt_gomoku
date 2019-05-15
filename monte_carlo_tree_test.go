package main

import (
	"sort"
	"testing"
	"time"
)

func TestLookupPiece(t *testing.T) {
	game, err := NewGame(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer game.TearDown()
	node, err := game.mctRoot.Expand()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("node pos:", node.Pos)
	isFailed := true
	for p := MinPosition; p <= MaxPosition; p++ {
		if piece := node.LookupPiece(p); piece != 0 {
			t.Logf("node.LookupPiece(%v) = %v", p, piece)
			isFailed = false
		}
	}
	if isFailed {
		t.Fail()
	}
}

func TestNewMonteCarloTreeForNon0Step(t *testing.T) {
	game, err := NewGame(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer game.TearDown()
	h8, err := ParsePosition("h8")
	if err != nil {
		t.Fatal(err)
	}
	h9, err := ParsePosition("h9")
	if err != nil {
		t.Fatal(err)
	}
	game.updateHistoryAndBoard(h8)
	step := game.mctRoot.Step + 1
	root, err := NewMonteCarloTree(game, step, h9)
	if err != nil {
		t.Fatal(err)
	}
	if root.IsFullyExpanded() {
		t.Fail()
	}
	if root == nil {
		t.Fatal("new root is nil")
	}
	logMctNodeInfo(t, root)
}

func TestExpand(t *testing.T) {
	game, err := NewGame(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer game.TearDown()
	node, err := game.mctRoot.Expand()
	if err != nil {
		t.Fatal(err)
	}
	if node.IsFullyExpanded() {
		t.Fail()
	}
	if node == nil {
		t.Fatal("node is nil")
	}
	logRootInfo(t, game)
	logMctNodeInfo(t, node)
}

func TestRollout1(t *testing.T) {
	game, err := NewGame(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer game.TearDown()
	t.Log(game.mctRoot.Rollout())
}

func TestRollout2(t *testing.T) {
	game, err := NewGame(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer game.TearDown()
	node, err := game.mctRoot.Expand()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(node.Rollout())
}

func TestSimulate(t *testing.T) {
	game, err := NewGame(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer game.TearDown()
	et, err := game.mctRoot.Simulate()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Elapsed time:", et)
	logRootInfo(t, game)
}

func TestSimulate2Times(t *testing.T) {
	testSimulateNTimes(t, 2)
}

func TestSimulate5Times(t *testing.T) {
	testSimulateNTimes(t, 5)
}

func TestSimulate225Times(t *testing.T) {
	testSimulateNTimes(t, 225)
}

func TestGetBestUctChild(t *testing.T) {
	game, err := NewGame(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer game.TearDown()
	for i := 0; i < 225; i++ {
		_, err = game.mctRoot.Simulate()
		if err != nil {
			t.Fatal(err)
		}
	}
	logRootInfo(t, game)
	//fmt.Println("* Call GetBestUctChild():")
	bestUctChild := game.mctRoot.GetBestUctChild()
	logMctNodeInfo(t, bestUctChild)
}

func TestSimulate226Times(t *testing.T) {
	testSimulateNTimes(t, 226)
}

func TestSimulate300Times(t *testing.T) {
	testSimulateNTimes(t, 300)
}

func TestMonteCarloTreeSearch(t *testing.T) {
	game, err := NewGame(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer game.TearDown()
	bestChild, err := game.mctRoot.MonteCarloTreeSearch()
	if err != nil {
		t.Fatal(err)
	}
	logRootInfo(t, game)
	if bestChild == nil {
		t.Log("Best child is nil.")
		return
	}
	t.Log("Best child pos:", bestChild.Pos)
}

func testSimulateNTimes(t *testing.T, n int) {
	game, err := NewGame(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer game.TearDown()
	var etSum time.Duration
	for i := 0; i < n; i++ {
		et, err := game.mctRoot.Simulate()
		if err != nil {
			t.Fatal(err)
		}
		etSum += et
	}
	t.Log("Total elapsed time:", etSum)
	t.Log("Average elasped time:", float64(etSum)/float64(n))
	logRootInfo(t, game)
}

func logRootInfo(tb testing.TB, game *Game) {
	logMctNodeInfo(tb, game.mctRoot)
}

func logMctNodeInfo(tb testing.TB, mctNode *MonteCarloTreeNode) {
	tb.Logf("Node - Pos: %v, NumWin: %d, NumSim: %d, UCT: %.6f",
		mctNode.Pos, mctNode.NumWin, mctNode.NumSim, mctNode.Uct())
	tb.Logf("Unexpanded pos: (len = %d) %v",
		len(mctNode.unexpPos), mctNode.unexpPos)
	sorted := append(mctNode.unexpPos[:0:0], mctNode.unexpPos...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	tb.Log("  Sorted:", sorted)
	tb.Log("Expanded nodes:")
	minPos := MaxPosition + 1
	maxPos := MinPosition - 1
	i := 0
	for node := mctNode.LastChild; node != nil; node = node.PrevSibling {
		if node.Pos < minPos {
			minPos = node.Pos
		}
		if node.Pos > maxPos {
			maxPos = node.Pos
		}
		i++
		tb.Logf("  %d - Pos: %v, NumWin: %d, NumSim: %d, UCT: %.6f",
			i, node.Pos, node.NumWin, node.NumSim, node.Uct())
	}
	if i > 0 {
		tb.Logf("  Pos min: %v, max: %v", minPos, maxPos)
	} else {
		tb.Log("  None")
	}
}
