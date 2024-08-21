package pwnpi

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/alecthomas/kong"
	"github.com/cmj0121/pwnpi/pkg/pidfile"
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
	Debug   bool             `short:"d" name:"debug" help:"Show the debug output."`

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

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	switch p.Debug {
	case true:
		ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
	default:
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		defer signal.Stop(sig)
		defer close(sig)

		select {
		case <-sig:
			log.Warn().Msg("Ctrl+C pressed, exiting ...")
			cancel()
		case <-ctx.Done():
		}
	}()

	return p.run(ctx, cmd)
}

func (p *PwnPi) run(ctx context.Context, cmd string) error {
	log.Info().Str("command", cmd).Msg("starting run pwnpi ...")
	defer log.Info().Msg("finished run pwnpi ...")

	lock := pidfile.New(PROJ_NAME)
	lock.Lock()
	defer lock.Release()

	switch cmd {
	case "pwn":
		log.Info().Msg("running pwn command ...")
		return p.Pwn.Run(ctx)
	case "install":
		log.Info().Msg("running install command ...")
		return p.Install.Run(ctx)
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
