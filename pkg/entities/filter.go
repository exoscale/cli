package entities

type ObjectFilterFunc func(ObjectInterface) bool

type ObjectVersionFilterFunc func(ObjectVersionInterface) bool

type UnversionedObject struct {
	*Object
}

func AsVersionFilter(f ObjectFilterFunc) ObjectVersionFilterFunc {
	return func(ov ObjectVersionInterface) bool {
		return f(ov)
	}
}

func AcceptAll(ObjectInterface) bool {
	return true
}
