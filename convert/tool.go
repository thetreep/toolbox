package convert

func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

type test interface {
	Fatalf(format string, args ...any)
}

func MustForTest[T any](v T, err error) func(t test) T {
	return func(t test) T {
		if err != nil {
			t.Fatalf("err: %+v", err)
		}
		return v
	}
}
