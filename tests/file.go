package tests

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/thetreep/toolbox/random"
)

// CreateTmpFile is a small helper for creating a tmp file.
func CreateTmpFile() (*os.File, error) {
	file, err := ioutil.TempFile(os.TempDir(), random.NewIDGenerator().NewID())
	if err != nil {
		return nil, errors.Wrap(err, "an error occurred while creating a tmp file")
	}

	return file, nil
}

// Setup sets the context for testing.
func Setup(t *testing.T, callback func(context.Context)) {
	t.Parallel()

	ctx := random.NewContext(
		context.Background(),
		random.NewIDGenerator(),
	)

	timeout := 4 * time.Minute
	ctx, cancel := context.WithTimeout(ctx, timeout)

	defer cancel()

	callback(ctx)
}

// CreateTmpFileWithContent creates a tmp file with the given content.
func CreateTmpFileWithContent(content string) (*os.File, error) {
	file, err := ioutil.TempFile(os.TempDir(), "tmp")
	if err != nil {
		return nil, errors.Wrap(err, "an error occurred while creating a tmp file")
	}

	err = ioutil.WriteFile(file.Name(), []byte(content), 0600)
	if err != nil {
		return nil, errors.Wrap(err, "an error occurred while writing to a tmp file")
	}

	err = file.Close()
	if err != nil {
		return nil, errors.Wrap(err, "an error occurred while closing a tmp file")
	}

	return file, nil
}
