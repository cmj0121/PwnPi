// Run the pwn tool with the given configuration.
package pwnpi

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/cmj0121/pwnpi/pkg/waveshare"
	"github.com/rs/zerolog/log"
)

//go:embed assets/**/*
var fs embed.FS

type Action int

const (
	IDLE Action = iota
	SLEEP
	CLOCK
)

const (
	// Duration from idle to sleep mode
	TO_IDLE_DURATION       = 180 * time.Second
	IDLE_TO_SLEEP_DURATION = 10 * time.Second
)

// The Pwn instance that control the PwnPi CLI and how it behaves.
type Pwn struct {
	// The flush interval of the display
	Interval time.Duration `name:"interval" default:"10ms" help:"The flush interval of the display."`

	// The waveshare E-Ink display instance
	display   *waveshare.WaveShare
	action    Action
	activated time.Time
}

// Run the PwnPi CLI based on the current configuration
func (p *Pwn) Run(ctx context.Context) (err error) {
	err = errors.Join(err, p.prologue())
	defer p.epilogue()

	err = errors.Join(err, p.display.Initialize())
	p.action = CLOCK
	p.activated = time.Now()

	return p.run(ctx)
}

func (p *Pwn) run(ctx context.Context) error {
	ticker := time.NewTicker(p.Interval)

	for {
		select {
		case <-ctx.Done():
			log.Warn().Msg("context is done, stop running the PwnPi CLI")
			return nil
		case <-ticker.C:
			p.handleAction()
		}
	}
}

func (p *Pwn) handleAction() {
	switch p.action {
	case SLEEP:
		// already in sleep mode, skip the sleep mode
	case IDLE:
		// only refresh the IDLE screen at first 5 seconds in IDLE mode
		if time.Since(p.activated) > IDLE_TO_SLEEP_DURATION {
			p.action = SLEEP
			if err := p.display.DeepSleep(); err != nil {
				log.Warn().Err(err).Msg("failed to enter the sleep mode")
			}
		}

		if err := p.display.ShowText("PwnPi", true, 240, 120, 64); err != nil {
			log.Warn().Err(err).Msg("failed to show the idle screen")
		}
	case CLOCK:
		now := time.Now().Format("15:04")
		if err := p.display.ShowText(now, true, 240, 120, 64); err != nil {
			log.Warn().Err(err).Msg("failed to show the clock screen")
		}

		if p.action != SLEEP && p.activated.Add(TO_IDLE_DURATION).Before(time.Now()) {
			log.Info().Msg("enter the idle mode")
			p.action = IDLE
		}
	}
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

	log.Info().Msg("PwnPi CLI is terminated")
	fmt.Println("~ Bye ~")
}
