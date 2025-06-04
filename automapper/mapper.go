package automapper

type MapFunc[T any, U any] func(T) U

func (a MapFunc[T, U]) Map(v T) U {
	return a(v)
}

func (a MapFunc[T, U]) MapEach(v []T) []U {
	var values []U //nolint:prealloc // need to keep nil value
	for _, value := range v {
		values = append(values, a.Map(value))
	}

	return values
}
