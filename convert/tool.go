package convert

func Must[T any](value T, err error) T {
	panicIfErr(err)
	return value
}

func PanicIfErrV[T any](v T, err error) T {
	panicIfErr(err)
	return v
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
