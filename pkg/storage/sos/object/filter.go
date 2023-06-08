package object

type ObjectFilterFunc func(ObjectInterface) bool

type ObjectVersionFilterFunc func(ObjectVersionInterface) bool

func AsVersionFilter(f ObjectFilterFunc) ObjectVersionFilterFunc {
	return func(ov ObjectVersionInterface) bool {
		return f(ov)
	}
}
