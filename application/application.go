package application

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli"
	"go.opencensus.io/trace"

	"github.com/thetreep/toolbox/binary"
	"github.com/thetreep/toolbox/process"
	"github.com/thetreep/toolbox/random"
)

// Application represents the cli application which will run when the binary is run.
type Application struct {
	*cli.App
}

const (
	// AppAuthor is the app's author.
	AppAuthor = "The Treep Backend Team"

	// AppCopyright specifies the app's copyright.
	AppCopyright = "TheTreep"

	// EnvironmentCliFlag specifies the environment in which the app runs.
	EnvironmentCliFlag = "THETREEP_ENV"

	// DefaultEnvironment is the default environment in which the app runs.
	DefaultEnvironment = "dev"
)

// NewApplication returns a new Application ready to `Run`.
func NewApplication(config Config, run func(context.Context, Info) error) *Application { // nolint
	binary.SetBinaryMetadata(config.Metadata)

	version := fmt.Sprintf(
		"Branch: %s,\n	 Compiler: %s,\n	 CompiledAt: %s,\n	 Commit: %s",
		config.Metadata.Branch,
		config.Metadata.Compiler,
		config.Metadata.CompiledAt,
		config.Metadata.Sha,
	)

	flags := []cli.Flag{EnvironmentFlag()}
	for i := range config.AdditionalFlags {
		flags = append(flags, &config.AdditionalFlags[i])
	}

	return &Application{
		&cli.App{
			Name:      config.Name,
			Author:    AppAuthor,
			Copyright: AppCopyright,

			EnableBashCompletion: true,

			UsageText: config.Usage,
			Usage:     config.Usage,

			Version: version,

			Metadata: config.Metadata.ToMap(),

			Flags: flags,

			Action: func(c *cli.Context) error {
				values := map[string]string{}

				for i := range config.AdditionalFlags {
					values[config.AdditionalFlags[i].Name] = GetFlagFromContext(
						c,
						config.AdditionalFlags[i].Name,
					)
				}

				info := Info{
					Environment: GetEnvironment(c),
					Values:      values,
				}

				ctx := random.NewContext(
					context.Background(),
					random.NewIDGenerator(),
				)

				ctx = process.NewContext(
					ctx,
					2*time.Second,
				)

				_, span := trace.StartSpan(ctx, "main")
				span.AddAttributes(trace.StringAttribute("version", version))
				defer span.End()

				return run(ctx, info)
			},
		},
	}
}

// EnvironmentFlag creates a Flag for the environment handling.
func EnvironmentFlag() cli.StringFlag {
	return cli.StringFlag{
		Name:   "environment",
		Usage:  "environment in which the app runs (production, dev, staging)",
		EnvVar: EnvironmentCliFlag,
		Value:  DefaultEnvironment,
	}
}

// GetEnvironment returns environment.
func GetEnvironment(c *cli.Context) string {
	return c.GlobalString("environment")
}

// GetFlagFromContext returns a flag value.
func GetFlagFromContext(c *cli.Context, flagName string) string {
	return c.GlobalString(flagName)
}
