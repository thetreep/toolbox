package process

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"
)

type key int

const (
	timeoutKey key = iota
)

// NewContext added a timeout to the Context.
func NewContext(ctx context.Context, timeout time.Duration) context.Context {
	return context.WithValue(ctx, timeoutKey, timeout)
}

// FromContext retrieves a timeout value for context.
func FromContext(ctx context.Context) time.Duration {
	timeout := 2 * time.Second

	t, ok := ctx.Value(timeoutKey).(time.Duration)
	if !ok || t < 100*time.Millisecond {
		return timeout
	}

	return t
}

// Process interface.
type Process interface {
	Run() error
	Close(ctx context.Context) error
}

// Run runs given server in parallel.
func Run(ctx context.Context, processes ...Process) error {
	if len(processes) == 0 {
		return errors.New("no server to run")
	}

	runner, ctx := errgroup.WithContext(ctx)

	for i := range processes {
		server := processes[i]
		runner.Go(server.Run)
	}

	errChan := make(chan error, 1)
	defer close(errChan)

	go func() {
		err := runner.Wait()
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	time.Sleep(FromContext(ctx))
	select {
	case err := <-errChan:
		if err != nil {
			return err
		}

	default:
		signalChan := make(chan os.Signal, 1)
		defer close(signalChan)

		signal.Notify(
			signalChan,
			syscall.SIGHUP,  // kill -SIGHUP XXXX
			syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
			syscall.SIGQUIT, // kill -SIGQUIT XXXX
		)

		<-signalChan

		go func() {
			<-signalChan
		}()

		closer, ctx := errgroup.WithContext(ctx)

		for i := range processes {
			server := processes[i]

			runner.Go(func() error {
				return server.Close(ctx)
			})
		}

		return closer.Wait()
	}

	return nil
}
