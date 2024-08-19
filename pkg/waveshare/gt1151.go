// GT1151 controller
package waveshare

import (
	"errors"
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

// The I2C controller for the GT1151.
type I2C struct {
	i2c.BusCloser
	i2c.Dev
}

func NewI2C() (*I2C, error) {
	bus, err := i2creg.Open("")
	if err != nil {
		log.Debug().Err(err).Msg("failed to open the I2C bus")
		return nil, err
	}

	i2c := &I2C{
		BusCloser: bus,
		Dev:       i2c.Dev{Addr: 0x14, Bus: bus},
	}
	return i2c, nil
}

func (i *I2C) Close() {
	i.BusCloser.Close()
}

// Initialize the GT1151 controller.
func (i *I2C) Initialize(g *GPIO) (err error) {
	// reset the GT1151 controller by TRST pin
	err = errors.Join(err, g.TRST.Out(gpio.Low))
	time.Sleep(100 * time.Millisecond)
	err = errors.Join(err, g.TRST.Out(gpio.High))

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
