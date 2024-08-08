package loser_test

import (
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

func (it *List[E]) At() E {
	return it.cur
}

func (it *List[E]) Next() bool {
	if len(it.list) > 0 {
		it.cur = it.list[0]
		it.list = it.list[1:]
		return true
	}
	var zero E
	it.cur = zero
	return false
}

func (it *List[E]) Seek(val E) bool {
	for it.cur.Less(val) && len(it.list) > 0 {
		it.cur = it.list[0]
		it.list = it.list[1:]
	}
	return len(it.list) > 0
}

func checkIterablesEqual[E loser.Lesser[E], S1, S2 loser.Sequence[E]](t *testing.T, a S1, b S2) {
	t.Helper()
	count := 0
	for a.Next() {
		count++
		if !b.Next() {
			t.Fatalf("b ended before a after %d elements", count)
		}
		if a.At().Less(b.At()) || b.At().Less(a.At()) {
			t.Fatalf("position %d: %v != %v", count, a.At(), b.At())
		}
	}
	if b.Next() {
		t.Fatalf("a ended before b after %d elements", count)
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name string
		args []*List[Uint64]
		want *List[Uint64]
	}{
		{
			name: "empty input",
			want: NewList[Uint64](),
		},
		{
			name: "one list",
			args: []*List[Uint64]{NewList[Uint64](1, 2, 3, 4)},
			want: NewList[Uint64](1, 2, 3, 4),
		},
		{
			name: "two lists",
			args: []*List[Uint64]{NewList[Uint64](3, 4, 5), NewList[Uint64](1, 2)},
			want: NewList[Uint64](1, 2, 3, 4, 5),
		},
		{
			name: "two lists, first empty",
			args: []*List[Uint64]{NewList[Uint64](), NewList[Uint64](1, 2)},
			want: NewList[Uint64](1, 2),
		},
		{
			name: "two lists, second empty",
			args: []*List[Uint64]{NewList[Uint64](1, 2), NewList[Uint64]()},
			want: NewList[Uint64](1, 2),
		},
		{
			name: "two lists b",
			args: []*List[Uint64]{NewList[Uint64](1, 2), NewList[Uint64](3, 4, 5)},
			want: NewList[Uint64](1, 2, 3, 4, 5),
		},
		{
			name: "two lists c",
			args: []*List[Uint64]{NewList[Uint64](1, 3), NewList[Uint64](2, 4, 5)},
			want: NewList[Uint64](1, 2, 3, 4, 5),
		},
		{
			name: "three lists",
			args: []*List[Uint64]{NewList[Uint64](1, 3), NewList[Uint64](2, 4), NewList[Uint64](5)},
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
		lists = append(lists, NewList(items...))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset the Lists to their original values each time round the loop.
		for j := range data {
			*lists[j] = List[Uint64]{list: data[j]}
		}
		lt := loser.New(lists, math.MaxUint64)
		consume(lt)
	}
}

func consume[E loser.Lesser[E], S loser.Sequence[E]](t *loser.Tree[E, S]) {
	for t.Next() {
		t.At()
	}
}
