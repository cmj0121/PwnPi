package waveshare

import (
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"periph.io/x/conn/v3/gpio"
)

// full-reset the E-Ink display
func (w *WaveShare) fullReset() (err error) {
	err = errors.Join(err, w.GPIO.Reset())
	w.GPIO.WaitToIdle()

	err = errors.Join(err, w.sendCommand(SW_RESET))
	w.GPIO.WaitToIdle()

	err = errors.Join(err, w.sendCommand(DRIVER_OUTPUT_CTRL, 0xf9, 0x00, 0x00))
	err = errors.Join(err, w.sendCommand(DATA_ENTRY_MODE, 0x03))
	err = errors.Join(err, w.sendCommand(BORDER_WAVEFROM_CTRL, 0x05))
	err = errors.Join(err, w.sendCommand(DISPLAY_CONTROL_1, 0x00, 0x80))
	err = errors.Join(err, w.sendCommand(TEMP_SENSOR_CTRL, 0x80))
	w.GPIO.WaitToIdle()

	err = errors.Join(err, w.setWindow(TP2in13_WIDTH, TP2in13_HEIGHT))
	err = errors.Join(err, w.setCursor(0, 0))
	w.GPIO.WaitToIdle()

	log.Debug().Err(err).Msg("full reset the WaveShare E-Ink display")
	return
}

// partial-reset the E-Ink display
func (w *WaveShare) partialReset() (err error) {
	err = errors.Join(err, w.GPIO.RST.Out(gpio.Low))
	time.Sleep(1 * time.Millisecond)
	err = errors.Join(err, w.GPIO.RST.Out(gpio.High))

	err = errors.Join(err, w.sendCommand(DRIVER_OUTPUT_CTRL, 0xF9, 0x00, 0x00))
	err = errors.Join(err, w.sendCommand(BORDER_WAVEFROM_CTRL, 0x80))
	err = errors.Join(err, w.sendCommand(DATA_ENTRY_MODE, 0x03))

	err = errors.Join(err, w.setWindow(TP2in13_WIDTH, TP2in13_HEIGHT))
	err = errors.Join(err, w.setCursor(0, 0))

	log.Debug().Err(err).Msg("partial reset the WaveShare E-Ink display")
	return
}

// Refresh the WaveShare E-Ink display.
func (w *WaveShare) refreshDisplay(fast bool) (err error) {
	var mode byte
	switch fast {
	case true:
		mode = 0xFF
	default:
		mode = 0xF7
	}

	err = errors.Join(err, w.sendCommand(DISPLAY_CONTROL_2, mode))
	err = errors.Join(err, w.sendCommand(MASTER_ACTIVATION))

	w.GPIO.WaitToIdle()

	log.Debug().Err(err).Msg("refreshing the WaveShare E-Ink display")
	return err
}

// Send the command to the WaveShare E-Ink display.
func (w *WaveShare) sendCommand(cmd Command, data ...byte) (err error) {
	err = errors.Join(err, w.DC.Out(gpio.Low))
	err = errors.Join(err, w.SPI.Write(byte(cmd)))

	if err == nil && len(data) > 0 {
		err = errors.Join(err, w.DC.Out(gpio.High))
		err = errors.Join(err, w.SPI.Write(data...))
	}

	log.Debug().Err(err).Str("command", cmd.String()).Msg("send the command")
	return err
}

// Set the window size of the WaveShare E-Ink display.
func (w *WaveShare) setWindow(width, height int) (err error) {
	width, height = width-1, height-1

	err = errors.Join(err, w.sendCommand(SET_RAM_X, 0, byte((width>>3)&0xFF)))
	err = errors.Join(err, w.sendCommand(SET_RAM_Y, 0, 0, byte(height&0xFF), byte((height>>8)&0xFF)))

	log.Debug().Err(err).Msg("set the window size")
	return err
}

// Set the cursor position of the WaveShare E-Ink display.
func (w *WaveShare) setCursor(x, y int) (err error) {
	err = errors.Join(err, w.sendCommand(SET_RAM_X_COUNTER, byte(x&0xFF)))
	err = errors.Join(err, w.sendCommand(SET_RAM_Y_COUNTER, byte(y&0xFF), byte((y>>8)&0xFF)))

	log.Debug().Err(err).Msg("set the cursor position")
	return err
}

// Show the image per bit-plane on the WaveShare E-Ink display.
func (w *WaveShare) showPixel(fast bool, img ...byte) (err error) {
	switch fast {
	case true:
		err = errors.Join(err, w.partialReset())
	default:
		err = errors.Join(err, w.fullReset())
	}

	err = errors.Join(err, w.sendCommand(WRITE_MEMORY_B_W, img...))
	err = errors.Join(err, w.sendCommand(NOP))
	err = errors.Join(err, w.refreshDisplay(fast))

	w.GPIO.WaitToIdle()

	log.Debug().Err(err).Msg("showing the image on the WaveShare E-Ink display")
	return err
}
