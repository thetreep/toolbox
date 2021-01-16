package application

import (
	"context"
	"fmt"

	"github.com/urfave/cli"
	"go.opencensus.io/trace"

	"github.com/thetreep/toolbox/binary"
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

	// DefaultConfigurationPath specifies the default path of the configuration file.
	DefaultConfigurationPath = "configuration.yml"

	// ConfigurationCliFlag specifies the path of the configuration file.
	ConfigurationCliFlag = "THETREEP_CONF_FILE"
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

	flags := []cli.Flag{
		EnvironmentFlag(),
		ConfigurationFlag(),
	}

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
					Environment:       GetEnvironment(c),
					ConfigurationPath: GetConfigPathFromContext(c),
					Values:            values,
				}

				ctx := context.Background()

				_, span := trace.StartSpan(ctx, config.Name)
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

// GetConfigPathFromContext reads the configuration file path from context.
func GetConfigPathFromContext(ctx *cli.Context) string {
	path := ctx.GlobalString("configuration")
	if path == "" {
		return DefaultConfigurationPath
	}

	return path
}

// ConfigurationFlag creates a Flag for the cli app.
func ConfigurationFlag() cli.StringFlag {
	return cli.StringFlag{
		Name:   "configuration",
		Usage:  "path to configuration file (yml)",
		EnvVar: ConfigurationCliFlag,
		Value:  DefaultConfigurationPath,
	}
}

// GetFlagFromContext returns a flag value.
func GetFlagFromContext(c *cli.Context, flagName string) string {
	return c.GlobalString(flagName)
}
