// Run the pwn tool with the given configuration.
package pwnpi

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"sync"
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
	IDLE_TO_SLEEP_DURATION = 60 * time.Second
)

// The Pwn instance that control the PwnPi CLI and how it behaves.
type Pwn struct {
	// The flush interval of the display
	Interval time.Duration `name:"interval" default:"10ms" help:"The flush interval of the display."`

	mu sync.Mutex

	// The waveshare E-Ink display instance
	display   *waveshare.WaveShare
	action    Action
	activated time.Time
	updating  bool
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

// The incoming I/O handler that serve the touch event and update the display.
func (p *Pwn) run(ctx context.Context) error {
	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()

	userEvent := p.display.Event(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("context is done, stop running the PwnPi CLI")
			return nil
		case event := <-userEvent:
			log.Info().Str("event", event.String()).Msg("user event received")
			p.activated = time.Now()

			switch p.action {
			case SLEEP, IDLE:
				log.Info().Msg("wake up the display")
				p.action = CLOCK
			}
		case <-ticker.C:
			// log.Trace().Msg("refresh the display ...")
			if !p.updating {
				// refresh the display in another goroutine, and it may be ignored
				// if another updating is running.
				go p.updateDisplay(p.action)
			}
		}
	}
}

// Update the display by the given action.
func (p *Pwn) updateDisplay(action Action) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.updating {
		log.Warn().Msg("still updating the display, skip the update")
		return
	}

	p.updating = true
	defer func() {
		p.updating = false
	}()

	switch action {
	case SLEEP:
		// already in sleep mode, skip the sleep mode
	case IDLE:
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
