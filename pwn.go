// Run the pwn tool with the given configuration.
package pwnpi

import (
	"context"
	"embed"
	"errors"

	"github.com/cmj0121/pwnpi/pkg/waveshare"
	"github.com/rs/zerolog/log"
)

//go:embed assets/**/*
var fs embed.FS

// The Pwn instance that control the PwnPi CLI and how it behaves.
type Pwn struct {
	// The waveshare E-Ink display instance
	display *waveshare.WaveShare
}

// Run the PwnPi CLI based on the current configuration
func (p *Pwn) Run(ctx context.Context) (err error) {
	err = errors.Join(err, p.prologue())
	defer p.epilogue()

	err = errors.Join(err, p.display.Initialize())

	for event := range p.display.Touches(ctx) {
		log.Info().Msgf("event: %v", event)
	}

	return
}

func (p *Pwn) prologue() error {
	display, err := waveshare.New(fs)
	if err != nil {
		log.Warn().Err(err).Msg("failed to create the waveshare E-Ink display")
		return err
	}

	p.display = display

	return nil
}

func (p *Pwn) epilogue() {
	if p.display != nil {
		p.display.Close()
	}
}
