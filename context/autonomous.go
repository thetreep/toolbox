package context

import (
	"context"
	"time"
)

type autonomous struct {
	ctx context.Context
}

func (c autonomous) Deadline() (time.Time, bool)       { return time.Time{}, false }
func (c autonomous) Done() <-chan struct{}             { return nil }
func (c autonomous) Err() error                        { return nil }
func (c autonomous) Value(key interface{}) interface{} { return c.ctx.Value(key) }

// WithoutCancel returns a context that is never canceled.
func Autonomous(ctx context.Context) context.Context {
	return autonomous{ctx: ctx}
}
