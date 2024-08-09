package loser_test

import (
	"iter"
	"math"
	"testing"

	"github.com/bboreham/go-loser"
)

type List[E loser.Lesser[E]] struct {
	list []E
	cur  E
}

func NewList[E loser.Lesser[E]](list ...E) *List[E] {
	return &List[E]{list: list}
}

func (it *List[E]) Iter() iter.Seq[E] {
	return func(yield func(E) bool) {
		for _, i := range it.list {
			yield(i)
		}
	}
}

func checkIterablesEqual[E loser.Lesser[E]](t *testing.T, a loser.Iterable[E], b loser.Iterable[E]) {
	t.Helper()
	count := 0
	next, stop := iter.Pull(b.Iter())
	defer stop()
	for va := range a.Iter() {
		count++
		vb, ok := next()
		if !ok {
			t.Fatalf("b ended before a after %d elements", count)
		}
		if va.Less(vb) || vb.Less(va) {
			t.Fatalf("position %d: %v != %v", count, va, vb)
		}
	}
	if _, ok := next(); ok {
		t.Fatalf("a ended before b after %d elements", count)
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name string
		args []loser.Iterable[Uint64]
		want *List[Uint64]
	}{
		{
			name: "empty input",
			want: NewList[Uint64](),
		},
		{
			name: "one list",
			args: []loser.Iterable[Uint64]{NewList[Uint64](1, 2, 3, 4)},
			want: NewList[Uint64](1, 2, 3, 4),
		},
		{
			name: "two lists",
			args: []loser.Iterable[Uint64]{NewList[Uint64](3, 4, 5), NewList[Uint64](1, 2)},
			want: NewList[Uint64](1, 2, 3, 4, 5),
		},
		{
			name: "two lists, first empty",
			args: []loser.Iterable[Uint64]{NewList[Uint64](), NewList[Uint64](1, 2)},
			want: NewList[Uint64](1, 2),
		},
		{
			name: "two lists, second empty",
			args: []loser.Iterable[Uint64]{NewList[Uint64](1, 2), NewList[Uint64]()},
			want: NewList[Uint64](1, 2),
		},
		{
			name: "two lists b",
			args: []loser.Iterable[Uint64]{NewList[Uint64](1, 2), NewList[Uint64](3, 4, 5)},
			want: NewList[Uint64](1, 2, 3, 4, 5),
		},
		{
			name: "two lists c",
			args: []loser.Iterable[Uint64]{NewList[Uint64](1, 3), NewList[Uint64](2, 4, 5)},
			want: NewList[Uint64](1, 2, 3, 4, 5),
		},
		{
			name: "three lists",
			args: []loser.Iterable[Uint64]{NewList[Uint64](1, 3), NewList[Uint64](2, 4), NewList[Uint64](5)},
			want: NewList[Uint64](1, 2, 3, 4, 5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := loser.New[Uint64](tt.args, math.MaxUint64)
			checkIterablesEqual(t, tt.want, lt)
		})
	}
}

type Uint64 uint64

func (u Uint64) Less(other Uint64) bool {
	return u < other
}

func BenchmarkMerge(b *testing.B) {
	var lists []*List[Uint64]
	var iterables []loser.Iterable[Uint64]
	var data [][]Uint64

	// Create 10000 Lists, so that all memory allocation is done before starting the benchmark.
	const nLists = 10000
	const nItems = 100
	for i := 0; i < nLists; i++ {
		items := make([]Uint64, 0, nItems)
		for j := 1; j < nItems; j++ {
			items = append(items, Uint64(i+j*nItems))
		}
		data = append(data, items)
		list := NewList(items...)
		lists = append(lists, list)
		iterables = append(iterables, list)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset the Lists to their original values each time round the loop.
		for j := range data {
			*lists[j] = List[Uint64]{list: data[j]}
		}
		lt := loser.New(iterables, math.MaxUint64)
		consume(lt)
	}
}

func consume[E loser.Lesser[E]](t *loser.Tree[E]) {
	for v := range t.Iter() {
		_ = v
	}
}
