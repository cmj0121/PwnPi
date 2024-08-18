// Run the pwn tool with the given configuration.
package pwnpi

import (
	"context"
	"errors"

	"github.com/cmj0121/pwnpi/pkg/waveshare"
	"github.com/rs/zerolog/log"
)

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

	return
}

func (p *Pwn) prologue() error {
	display, err := waveshare.New()
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
