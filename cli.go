package pwnpi

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// The project name of the PwnPi CLI
	PROJ_NAME = "pwnpi"
	// The version of the PwnPi CLI
	MAJOR = 0
	MINOR = 1
	PATCH = 0
)

// The PwnPi instance that control the PwnPi CLI and how it behaves.
type PwnPi struct {
	// general options
	Version kong.VersionFlag `short:"V" name:"version" help:"Show version information and exit."`

	// logging options
	Verbose int `short:"v" name:"verbose" type:"counter" help:"Show the verbose output."`

	// The command that the PwnPi CLI can execute
	Install *Install `cmd:"" help:"Install the PwnPi CLI into the Raspberry Pi."`
	Pwn     *Pwn     `cmd:"" help:"Run the PwnPi CLI."`
}

// Create a new PwnPi instance with default configuration
func New() *PwnPi {
	return &PwnPi{}
}

// Parse the command line arguments and run the PwnPi CLI
func (p *PwnPi) ParseAndRun() {
	opts := []kong.Option{
		kong.Name("pwnpi"),
		kong.Description("The PwnPi CLI for Raspberry Pi."),
		kong.UsageOnError(),
		kong.Vars{
			"version": fmt.Sprintf("%s (%d.%d.%d)", PROJ_NAME, MAJOR, MINOR, PATCH),
		},
	}

	ctx := kong.Parse(p, opts...)
	ctx.FatalIfErrorf(p.Run(ctx.Command()))
}

// Run the PwnPi CLI based on the current configuration
func (p *PwnPi) Run(cmd string) error {
	p.prologue()
	defer p.epilogue()

	return p.run(cmd)
}

func (p *PwnPi) run(cmd string) error {
	log.Info().Str("command", cmd).Msg("starting run pwnpi ...")
	defer log.Info().Msg("finished run pwnpi ...")

	switch cmd {
	case "pwn":
		log.Info().Msg("running pwn command ...")
		return p.Pwn.Run()
	case "install":
		log.Info().Msg("running install command ...")
		return p.Install.Run()
	}

	return nil
}

func (p *PwnPi) prologue() {
	// setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	writer := zerolog.ConsoleWriter{Out: os.Stderr}
	log.Logger = zerolog.New(writer).With().Timestamp().Logger()

	switch p.Verbose {
	case 0:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case 1:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case 2:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case 3:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	log.Debug().Msg("finished set up pwnpi ...")
}

func (p *PwnPi) epilogue() {
	log.Debug().Msg("starting clean up pwnpi ...")
	log.Debug().Msg("finished clean up pwnpi ...")
}
