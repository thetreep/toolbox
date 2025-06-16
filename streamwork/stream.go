package streamwork

//go:generate go run ./internal/gen.go

type streamConfig struct {
	continueOnError bool
	bufferSize      int
	errHandler      ErrHandler
}

type StreamOption func(*streamConfig)

func StreamOptionContinueOnError() StreamOption {
	return func(c *streamConfig) {
		c.continueOnError = true
	}
}

func StreamOptionBufferSize(size int) StreamOption {
	return func(c *streamConfig) {
		c.bufferSize = size
	}
}

type streamSink[T any] struct {
	c    <-chan T
	errG ErrGetter
}

func (s streamSink[T]) Collect() ([]T, error) {
	var result []T
	for v := range s.c {
		result = append(result, v)
	}
	return result, s.errG()
}

func stream[T any](fn func(cfg streamConfig) <-chan T, options ...StreamOption) ([]T, error) {
	cfg := streamConfig{}
	for _, opt := range options {
		opt(&cfg)
	}
	errC := &errCollector{}
	cfg.errHandler = errC.Handle
	return streamSink[T]{
		c:    fn(cfg),
		errG: errC.Err,
	}.Collect()
}
