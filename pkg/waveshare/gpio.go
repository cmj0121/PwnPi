package waveshare

import (
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3/rpi"
)

// The GPIO pin that the Waveshare e-Paper display is connected to.
type GPIO struct {
	RST  gpio.PinIO
	BUSY gpio.PinIO
	CS   gpio.PinIO
	DC   gpio.PinIO
	TRST gpio.PinIO
	INT  gpio.PinIO
}

func NewGPIO() (*GPIO, error) {
	gpio := &GPIO{
		RST:  rpi.P1_11,
		BUSY: rpi.P1_18,
		CS:   rpi.P1_24,
		DC:   rpi.P1_22,
		TRST: rpi.P1_15,
		INT:  rpi.P1_13,
	}

	log.Debug().Msg("successfully connected to the GPIO pins")
	return gpio, nil
}

func (g *GPIO) Close() {
	log.Debug().Msg("successfully closed the GPIO pins")
}

// Reset the GPIO pin.
func (g *GPIO) Reset() (err error) {
	err = errors.Join(err, g.RST.Out(gpio.Low))
	time.Sleep(2 * time.Millisecond)
	err = errors.Join(err, g.RST.Out(gpio.High))
	time.Sleep(20 * time.Millisecond)

	log.Debug().Err(err).Msg("reset the GPIO pin")
	return err
}

// Wait until the BUSY pin is low.
func (g *GPIO) WaitToIdle() {
	log.Debug().Msg("waiting for the GPIO pins to be idle")

	for g.BUSY.Read() == gpio.High {
		time.Sleep(1 * time.Millisecond)
		// log.Trace().Msg("waiting for the GPIO pins to be idle")
	}

	log.Debug().Msg("the GPIO pins are now idle")
}
