package wheel

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](values ...T) Set[T] {
	s := make(Set[T], len(values))
	s.Add(values...)
	return s
}

// Add appends new elements with specified values to the [Set].
func (s Set[T]) Add(values ...T) {
	for _, item := range values {
		s[item] = struct{}{}
	}
}

// Union returns a new [Set] containing all the elements in this [Set].
func (s Set[T]) Union(others ...Set[T]) Set[T] {
	length := len(s)
	for _, other := range others {
		length += len(other)
	}

	union := make(Set[T], length)
	union.Add(s.Values()...)
	for _, other := range others {
		union.Add(other.Values()...)
	}

	return union
}

// Values returns a slice of the values in the [Set].
func (s Set[T]) Values() []T {
	values := make([]T, 0, len(s))
	for value := range s {
		values = append(values, value)
	}

	return values
}
