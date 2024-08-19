package waveshare

import (
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3"
)

const (
	// The WaveShare 2.13 inch E-Ink display
	TP2in13_WIDTH  = 122
	TP2in13_HEIGHT = 250
)

type Command byte

// The WaveShare E-Ink display command set.
const (
	DRIVER_OUTPUT_CTRL   Command = 0x01
	GATE_VOLTAGE_CTRL    Command = 0x03
	SRC_VOLTAGE_CTRL     Command = 0x04
	INIT_OTP_PROGRAM     Command = 0x08
	BOOSTER_SOFT_START   Command = 0x0C
	DEEP_SLEEP_MODE      Command = 0x10
	DATA_ENTRY_MODE      Command = 0x11
	SW_RESET             Command = 0x12
	HV_READY             Command = 0x14
	VCI_DETECT           Command = 0x15
	TEMP_SENSOR_CTRL     Command = 0x18
	TEMP_SENSOR_SEL      Command = 0x1A
	TEMP_SENSOR_READ     Command = 0x1B
	TEMP_SENSOR_WRITE    Command = 0x1C
	MASTER_ACTIVATION    Command = 0x20
	DISPLAY_CONTROL_1    Command = 0x21
	DISPLAY_CONTROL_2    Command = 0x22
	WRITE_MEMORY_B_W     Command = 0x24
	WRITE_MEMORY_RED     Command = 0x26
	READ_MEMORY          Command = 0x27
	VCOM_SENSE           Command = 0x28
	VCOM_SENSE_READ      Command = 0x29
	VCOM_OTP_PROG        Command = 0x2A
	VCOM_SENSE_WRITE     Command = 0x2B
	USER_ID_READ         Command = 0x2E
	STATUS_READ          Command = 0x2F
	CRC_CALCULATION      Command = 0x34
	CRC_READ             Command = 0x35
	OPT_PROG_MODE        Command = 0x39
	BORDER_WAVEFROM_CTRL Command = 0x3C
	EOPT                 Command = 0x3F
	READ_RAM_OPT         Command = 0x41
	SET_RAM_X            Command = 0x44
	SET_RAM_Y            Command = 0x45
	AUTO_WRITE_RED       Command = 0x46
	AUTO_WRITE_B_W       Command = 0x47
	SET_RAM_X_COUNTER    Command = 0x4E
	SET_RAM_Y_COUNTER    Command = 0x4F
	NOP                  Command = 0x7F
)

func (c Command) String() string {
	return fmt.Sprintf("0x%02X", byte(c))
}

// The waveshare E-Ink display instance that control the display and how it behaves.
type WaveShare struct {
	*SPI
	*GPIO
	*I2C

	fs embed.FS
}

func New(fs embed.FS) (*WaveShare, error) {
	_, err := host.Init()
	if err != nil {
		log.Warn().Err(err).Msg("failed to initialize the host")
	}

	spi, err := NewSPI()
	if err != nil {
		log.Warn().Err(err).Msg("failed to create the SPI bus")
		return nil, err
	}

	gpio, err := NewGPIO()
	if err != nil {
		defer spi.Close()

		log.Warn().Err(err).Msg("failed to create the GPIO bus")
		return nil, err
	}

	i2c, err := NewI2C()
	if err != nil {
		defer spi.Close()
		defer gpio.Close()

		log.Warn().Err(err).Msg("failed to create the I2C bus")
		return nil, err
	}

	wave := &WaveShare{
		SPI:  spi,
		GPIO: gpio,
		I2C:  i2c,

		fs: fs,
	}

	return wave, nil
}

func (w *WaveShare) Close() {
	w.GPIO.Close()
	w.SPI.Close()
}

// Initialize the WaveShare E-Ink display.
//
// It will send the initialization commands to the display, reset the GPIO pins, and clear
// the display.
func (w *WaveShare) Initialize() (err error) {
	// initialize the SPI bus
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

	// initialize the I2C bus
	err = errors.Join(err, w.I2C.Initialize(w.GPIO))

	// initialize the display
	err = errors.Join(err, w.eraseDisplay(false))
	time.Sleep(333 * time.Millisecond)
	err = errors.Join(err, w.showText("PwnPi", 240, 120, 64))
	time.Sleep(333 * time.Millisecond)
	err = errors.Join(err, w.showAssets(IMG_RPI_ICON, 8))

	log.Info().Err(err).Msg("successfully initialized the WaveShare E-Ink display")
	return
}

// Erase and fill the bit color on the WaveShare E-Ink display.
//
// 0 bit is black, 1 bit is white.
func (w *WaveShare) eraseDisplay(fast bool) (err error) {
	log.Info().Msg("clearing the WaveShare E-Ink display")

	width, height := TP2in13_WIDTH, TP2in13_HEIGHT
	wire := (width + 7) / 8

	pixels := make([]byte, wire*height)
	pixels[0] = 0xFF
	for i := 1; i < len(pixels); i *= 2 {
		copy(pixels[i:], pixels[:i])
	}
	return w.showPixel(fast, pixels...)
}

// Show the image per bit-plane on the WaveShare E-Ink display.
func (w *WaveShare) showPixel(fast bool, img ...byte) (err error) {
	err = errors.Join(err, w.sendCommand(WRITE_MEMORY_B_W, img...))
	err = errors.Join(err, w.sendCommand(NOP))
	err = errors.Join(err, w.refreshDisplay(fast))

	log.Debug().Err(err).Msg("showing the image on the WaveShare E-Ink display")

	w.GPIO.WaitToIdle()
	return err
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
