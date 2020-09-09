package random

import (
	"context"
	"io"
	"math/rand"
	"time"

	"github.com/oklog/ulid"
)

type keyType string

const (
	key = keyType("id")
)

// Generator returns random ids with low rates of collisions.
type Generator interface {
	NewID() string
	Compare(string, string) int
}

// FromContext extracts an Generator from the Context.
func FromContext(ctx context.Context) Generator {
	return ctx.Value(key).(Generator)
}

// NewContext added an Generator to the Context.
func NewContext(ctx context.Context, id Generator) context.Context {
	return context.WithValue(ctx, key, id)
}

// NewIDGenerator creates an Generator.
func NewIDGenerator() Generator {
	t := time.Unix(1000000, 0)

	return &muon{
		entropy: rand.New(rand.NewSource(t.UnixNano())),
	}
}

// muon generates id via ulid.
type muon struct {
	entropy io.Reader
}

// NewID returns a new ID.
func (e *muon) NewID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), e.entropy).String()
}

// NewID compares IDs.
func (e *muon) Compare(a, b string) int {
	return ulid.MustParse(a).Compare(ulid.MustParse(b))
}

// FakeRandom returns a fake id generator.
func FakeRandom() Generator {
	return &fakeRandom{}
}

type fakeRandom struct{}

func (e *fakeRandom) NewID() string {
	return "2"
}

// NewID compares IDs.
func (e *fakeRandom) Compare(a, b string) int {
	if a == b {
		return 0
	}

	return 1
}
