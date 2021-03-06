package permute

import (
	"fmt"
	"reflect"
)

// fact is a factoradic encoding of an integer.
type fact []int

func newFact(i int64, n int) (fact, bool) {
	f := make(fact, n-1)
	for j := 2; j <= n; j++ {
		f[j-2] = int(i % int64(j))
		i /= int64(j)
	}

	return f, i == 0
}

func (f fact) perm() perm {
	p := identityPerm(len(f) + 1)

	e := len(f) - 1
	for i := range f {
		j := i + f[e-i]
		x := p[j]
		copy(p[i+1:j+1], p[i:j])
		p[i] = x
	}

	return p
}

// perm is a permutation of the numbers 0..n-1
type perm []int

func identityPerm(n int) perm {
	p := make(perm, n)
	for i := range p {
		p[i] = i
	}
	return p
}

func (p perm) apply(d Interface) {
	for i, j := range p {
		for j < i {
			j = p[j]
		}
		d.Swap(i, j)
	}
}

// next advances p to the next lexicographic ordering, and mirrors those swaps in d.
func (p perm) next(d Interface) bool {
	if len(p) != d.Len() {
		panic(fmt.Errorf("len(p) = %d, d.Len() = %d", len(p), d.Len()))
	}

	// Find the first number from the right that's smaller than its neighbor to the right.
	i := len(p) - 2
	for i >= 0 && p[i] >= p[i+1] {
		i--
	}

	// If the elements are in descending order.
	if i < 0 {
		return false
	}

	// Find the first number from the right that's larger than p[i].
	j := len(p) - 1
	for p[i] >= p[j] {
		j--
	}
	p[i], p[j] = p[j], p[i]
	d.Swap(i, j)

	// Reverse p[i+1:]
	for l, r := i+1, len(p)-1; l < r; l, r = l+1, r-1 {
		p[l], p[r] = p[r], p[l]
		d.Swap(l, r)
	}

	return true
}

func (p perm) inverse() perm {
	q := make(perm, len(p))
	for i, j := range p {
		q[j] = i
	}
	return q
}

type Interface interface {
	Len() int
	Swap(i, j int)
}

type Permuter struct {
	data  Interface
	p     perm
	first bool
	reset perm
}

func NewPermuter(data Interface) *Permuter {
	return &Permuter{
		data:  data,
		p:     identityPerm(data.Len()),
		first: true,
		reset: nil,
	}
}

type slice struct {
	s interface{}
}

func (s slice) Len() int {
	return reflect.ValueOf(s.s).Len()
}

func (s slice) Swap(i, j int) {
	reflect.Swapper(s.s)(i, j)
}

func NewSlicePermuter(data interface{}) *Permuter {
	return NewPermuter(slice{data})
}

// SetNext makes it so Permute yields the ith lexicographic permutation on its next call.
func (p *Permuter) SetNext(i int64) bool {
	f, ok := newFact(i, p.data.Len())
	if !ok {
		return false
	}

	if p.reset == nil {
		p.reset = p.p.inverse()
	}
	p.p = f.perm()
	p.first = true
	return true
}

// Permute updates the slice to the next permutation, and returns whether there are more permutations remaining.
func (p *Permuter) Permute() bool {
	if p.reset != nil {
		p.reset.apply(p.data)
		p.reset = nil
	}

	if p.first {
		p.first = false
		p.p.apply(p.data)
		return true
	}

	if p.p.next(p.data) {
		return true
	}

	// Set data back to its original order.
	p.p.inverse().apply(p.data)
	p.p = identityPerm(p.data.Len())
	return false
}
