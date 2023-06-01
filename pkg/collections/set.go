package collections

type empty struct{}

type Set[T comparable] map[T]empty

func NewSet[T comparable](values ...T) Set[T] {
	set := make(Set[T])
	for _, value := range values {
		set[value] = empty{}
	}

	return set
}

func (set Set[T]) Contains(v T) bool {
	_, ok := set[v]

	return ok
}
