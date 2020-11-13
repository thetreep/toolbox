package application

import (
	"github.com/thetreep/toolbox/binary"
	"github.com/urfave/cli"
)

// Config holds informations on the application about to run.
type Config struct {
	Metadata        binary.Metadata
	Name            string
	Usage           string
	AdditionalFlags []cli.StringFlag
}

// Info holds informations on the app.
type Info struct {
	ConfigurationPath string
	Environment       string
	Values            map[string]string
}
