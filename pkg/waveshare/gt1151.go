// GT1151 controller
package waveshare

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
)

const (
	GT_CTRL_REG  = 0x8040
	GT_MODSW_REG = 0x804D
	GT_CHECK_REG = 0x80FF
	GT_PID_REG   = 0x8140
	GT_GSTID_REG = 0x814E
	GT_TP1_REG   = 0x814F
	GT_TP2_REG   = 0x8157
	GT_TP3_REG   = 0x815F
	GT_TP4_REG   = 0x8167
	GT_TP5_REG   = 0x816F
)

// The touch event.
type Touch struct {
	TrackID int // the track ID

	X int // the x coordinate
	Y int // the y coordinate
	Z int // the pressure
}

func (t Touch) String() string {
	return fmt.Sprintf("(%d, %d) %d", t.X, t.Y, t.Z)
}

// The multi-point touch event
type MultiTouch struct {
	Count  int     // the number of touch points
	Touchs []Touch // the touch points
}

func (m MultiTouch) String() string {
	return fmt.Sprintf("%v", m.Touchs)
}

// The I2C controller for the GT1151.
type I2C struct {
	i2c.BusCloser
	i2c.Dev
	*GPIO

	// the handler for interrupt pin
	finish chan struct{}
}

func NewI2C(gpio *GPIO) (*I2C, error) {
	bus, err := i2creg.Open("")
	if err != nil {
		log.Debug().Err(err).Msg("failed to open the I2C bus")
		return nil, err
	}

	i2c := &I2C{
		BusCloser: bus,
		Dev:       i2c.Dev{Addr: 0x14, Bus: bus},
		GPIO:      gpio,
	}
	return i2c, nil
}

func (i *I2C) Close() {
	close(i.finish)
	i.BusCloser.Close()
}

// Initialize the GT1151 controller.
func (i *I2C) Initialize() (err error) {
	// reset the GT1151 controller by TRST pin
	err = errors.Join(err, i.TRST.Out(gpio.Low))
	time.Sleep(100 * time.Millisecond)
	err = errors.Join(err, i.TRST.Out(gpio.High))

	log.Info().Err(err).Msg("hard reset the GT1151 controller")

	// get the GT1151 controller ID
	switch id, err := i.readByte(GT_PID_REG, 4); err {
	case nil:
		log.Info().Bytes("id", id).Msg("successfully read the GT1151 controller ID")
	default:
		log.Warn().Err(err).Msg("failed to read the GT1151 controller ID")
	}

	return
}

// Read the byte from the GT1151 controller.
func (i *I2C) readByte(reg int, size int) ([]byte, error) {
	write := []byte{byte(reg >> 8), byte(reg & 0xFF)}
	read := make([]byte, size)

	err := i.Dev.Tx(write, read)
	if err != nil {
		log.Debug().Err(err).Msg("failed to read the byte from the GT1151 controller")
		return nil, err
	}

	return read, nil
}

// The iterator to read the touch event from the GT1151 controller.
func (i *I2C) Touches(ctx context.Context) <-chan MultiTouch {
	ch := make(chan MultiTouch)

	log.Info().Msg("start the GT1151 interrupt goroutine")
	defer log.Info().Msg("stop the GT1151 interrupt goroutine")

	go func() {
		defer close(ch)

		for {
			time.Sleep(1 * time.Millisecond)

			select {
			case <-ctx.Done():
				return
			case <-i.finish:
				return
			default:
				if i.INT.Read() != gpio.High {
					log.Trace().Msg("interrupt signal is low")
					continue
				}

				switch event := i.handleTouch(); event {
				case nil:
					log.Trace().Msg("failed to handle the touch event")
				default:
					ch <- *event
				}
			}
		}
	}()

	return ch
}

func (i *I2C) handleTouch() *MultiTouch {
	defer func() {
		write := []byte{byte(GT_GSTID_REG >> 8), byte(GT_GSTID_REG & 0xFF), 0x00}
		_, _ = i.Dev.Write(write)
	}()

	// receive the touch event
	buff, err := i.readByte(GT_GSTID_REG, 1)

	if err != nil || buff[0] == 0 {
		// log.Trace().Err(err).Msg("failed to read the touch event")
		return nil
	}

	// validate the event
	_, count := buff[0]&0x80, buff[0]&0x0F
	if count == 0 || count > 5 {
		// log.Trace().Int("count", int(count)).Msg("invalid touch count")
		return nil
	}

	// receive all the touch points
	buff, err = i.readByte(GT_TP1_REG, int(count)*8)
	if err != nil {
		log.Warn().Err(err).Msg("failed to read the touch points")
		return nil
	}

	event := &MultiTouch{Count: int(count)}
	for i := 0; i < int(count); i++ {
		touch := Touch{
			TrackID: int(buff[i*8+0]),
			X:       int(buff[i*8+1]) | int(buff[i*8+2])<<8,
			Y:       int(buff[i*8+3]) | int(buff[i*8+4])<<8,
			Z:       int(buff[i*8+5]) | int(buff[i*8+6])<<8,
		}

		event.Touchs = append(event.Touchs, touch)
	}

	return event
}
