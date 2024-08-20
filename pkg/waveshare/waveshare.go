package waveshare

import (
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
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

	i2c, err := NewI2C(gpio)
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
	err = errors.Join(err, w.fullReset())

	// initialize the I2C bus
	err = errors.Join(err, w.I2C.Initialize())

	// initialize the display
	err = errors.Join(err, w.showAssets(IMG_RPI_ICON, false, 8))
	time.Sleep(333 * time.Millisecond)
	err = errors.Join(err, w.ShowIdle())

	err = errors.Join(err, w.DeepSleep())

	log.Info().Err(err).Msg("successfully initialized the WaveShare E-Ink display")
	return
}

// Set the E-Ink display to deep sleep mode.
func (w *WaveShare) DeepSleep() (err error) {
	err = w.sendCommand(DEEP_SLEEP_MODE)
	w.WaitToIdle()

	log.Info().Err(err).Msg("set the WaveShare E-Ink display to deep sleep mode")
	return
}

func (w *WaveShare) ShowIdle() (err error) {
	err = errors.Join(err, w.showText("PwnPi", true, 240, 120, 64))
	return
}
