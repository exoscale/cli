package object

type ObjectFilterFunc func(ObjectInterface) bool

type ObjectVersionFilterFunc func(ObjectVersionInterface) bool

func ApplyFilters(obj ObjectInterface, filters []ObjectFilterFunc) bool {
	for _, filter := range filters {
		if !filter(obj) {
			return false
		}
	}

	return true
}

func ApplyVersionedFilters(obj ObjectVersionInterface, filters []ObjectVersionFilterFunc) bool {
	for _, filter := range filters {
		if !filter(obj) {
			return false
		}
	}

	return true
}
