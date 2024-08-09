// Loser tree, from https://en.wikipedia.org/wiki/K-way_merge_algorithm#Tournament_Tree

package loser

import (
	"iter"
)

type Lesser[T any] interface {
	Less(T) bool
}

type Iterable[E Lesser[E]] interface {
	Iter() iter.Seq[E]
}

func New[E Lesser[E]](sequences []Iterable[E], maxVal E) *Tree[E] {
	nSequences := len(sequences)
	t := Tree[E]{
		maxVal: maxVal,
		nodes:  make([]node[E], nSequences*2),
	}
	for i, s := range sequences {
		next, stop := iter.Pull(s.Iter())
		t.nodes[i+nSequences].next = next
		t.nodes[i+nSequences].stop = stop
		t.moveNext(i + nSequences) // Must call Next on each item so that At() has a value.
	}
	if nSequences > 0 {
		t.nodes[0].index = -1 // flag to be initialized on first call to Next().
	}
	return &t
}

// A loser tree is a binary tree laid out such that nodes N and N+1 have parent N/2.
// We store M leaf nodes in positions M...2M-1, and M-1 internal nodes in positions 1..M-1.
// Node 0 is a special node, containing the winner of the contest.
type Tree[E Lesser[E]] struct {
	maxVal E
	nodes  []node[E]
}

type node[E Lesser[E]] struct {
	index int              // This is the loser for all nodes except the 0th, where it is the winner.
	value E                // Value copied from the loser node, or winner for node 0.
	next  func() (E, bool) // Only populated for leaf nodes.
	stop  func()           // Only populated for leaf nodes.
}

func (t *Tree[E]) moveNext(index int) bool {
	n := &t.nodes[index]
	if v, ok := n.next(); ok {
		n.value = v
		return true
	}
	n.value = t.maxVal
	n.index = -1
	return false
}

func (t *Tree[E]) Iter() iter.Seq[E] {
	return func(yield func(E) bool) {
		nodes := t.nodes
		if len(nodes) == 0 {
			return
		}
		var cont bool
		t.initialize()
		if nodes[nodes[0].index].index != -1 {
			cont = yield(nodes[0].value)
		}
		for cont {
			t.moveNext(nodes[0].index)
			t.replayGames(nodes[0].index)
			if nodes[nodes[0].index].index == -1 {
				break
			}
			cont = yield(nodes[0].value)
		}
	}
}

func (t *Tree[E]) IsEmpty() bool {
	nodes := t.nodes
	if nodes[0].index == -1 { // If tree has not been initialized yet, do that.
		t.initialize()
	}
	return nodes[nodes[0].index].index == -1
}

func (t *Tree[E]) initialize() {
	winner := t.playGame(1)
	t.nodes[0].index = winner
	t.nodes[0].value = t.nodes[winner].value
}

// Find the winner at position pos; if it is a non-leaf node, store the loser.
// pos must be >= 1 and < len(t.nodes)
func (t *Tree[E]) playGame(pos int) int {
	nodes := t.nodes
	if pos >= len(nodes)/2 {
		return pos
	}
	left := t.playGame(pos * 2)
	right := t.playGame(pos*2 + 1)
	var loser, winner int
	if nodes[left].value.Less(nodes[right].value) {
		loser, winner = right, left
	} else {
		loser, winner = left, right
	}
	nodes[pos].index = loser
	nodes[pos].value = nodes[loser].value
	return winner
}

// Starting at pos, which is a winner, re-consider all values up to the root.
func (t *Tree[E]) replayGames(pos int) {
	nodes := t.nodes
	winningValue := nodes[pos].value
	for n := parent(pos); n != 0; n = parent(n) {
		node := &nodes[n]
		if node.value.Less(winningValue) {
			// Record pos as the loser here, and the old loser is the new winner.
			node.index, pos = pos, node.index
			node.value, winningValue = winningValue, node.value
		}
	}
	// pos is now the winner; store it in node 0.
	nodes[0].index = pos
	nodes[0].value = winningValue
}

func parent(i int) int { return i >> 1 }
