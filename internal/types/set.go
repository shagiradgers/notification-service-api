package types

type Set[K comparable] map[K]struct{}

func NewSet[K comparable]() Set[K] {
	return make(Set[K])
}

func (set Set[K]) Add(values ...K) {
	for _, v := range values {
		set[v] = struct{}{}
	}
}

func (set Set[K]) Merge(setToMerge Set[K]) Set[K] {
	for v := range setToMerge {
		set[v] = struct{}{}
	}
	return set
}

func (set Set[K]) ToSlice() []K {
	res := make([]K, 0, len(set))
	for v := range set {
		res = append(res, v)
	}
	return res
}
