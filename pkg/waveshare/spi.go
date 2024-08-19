package waveshare

import (
	"github.com/rs/zerolog/log"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
)

// The SPI bus that the Waveshare e-Paper display is connected to.
type SPI struct {
	port spi.PortCloser
	conn spi.Conn
}

// Connect to the SPI bus and return the SPI instance.
func NewSPI() (*SPI, error) {
	// find the first available SPI bus.
	port, err := spireg.Open("")
	if err != nil {
		log.Warn().Err(err).Msg("failed to open the SPI bus")
		return nil, err
	}

	// convert the SPI port to the SPI connection.
	conn, err := port.Connect(10*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		defer port.Close()

		log.Warn().Err(err).Msg("failed to connect to the SPI bus")
		return nil, err
	}

	spi := &SPI{
		port: port,
		conn: conn,
	}

	log.Debug().Msg("successfully connected to the SPI bus")
	return spi, nil
}

// Close the SPI bus connection.
func (s *SPI) Close() {
	s.port.Close()
	log.Debug().Msg("successfully closed the SPI bus connection")
}

// Write the data to the SPI bus.
func (s *SPI) Write(data ...byte) error {
	return s.conn.Tx(data, nil)
}
