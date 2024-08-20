// The PID file that ensure the process is running only once.
package pidfile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

const (
	DEFAULT_PID_FOLDER = "/var/run"
)

// The instance that hold the PID file information.
type PIDFile struct {
	name string
	path string
}

// Create a new PID file instance.
func New(name string) *PIDFile {
	return NewWithFolder(name, DEFAULT_PID_FOLDER)
}

// Create the PID file instance with the folder path.
func NewWithFolder(name, folder string) *PIDFile {
	path := fmt.Sprintf("%s/%s.pid", folder, name)
	abs, err := filepath.Abs(path)

	if err != nil {
		// cannot get the absolute path
		log.Fatal().Err(err).Str("path", path).Msg("failed to get the absolute path")
	}

	return &PIDFile{
		name: name,
		path: abs,
	}
}

// Get the PID file path.
func (p *PIDFile) Path() string {
	return p.path
}

// Lock the PID file instance and return error if the file is already locked.
func (p *PIDFile) TryLock() error {
	pid := fmt.Sprintf("%d", os.Getpid())

	switch _, err := os.Stat(p.path); err {
	case nil:
		// pid file exists, check the content
		data, err := os.ReadFile(p.path)
		if err != nil {
			log.Warn().Err(err).Str("path", p.path).Msg("failed to read the PID file")
			return err
		}

		if string(data) != pid {
			log.Warn().Str("path", p.path).Bytes("pid", data).Msg("the PID file is already locked")
			return fmt.Errorf("the PID file %s is already locked", p.name)
		}
	default:
		err = os.WriteFile(p.path, []byte(pid), 0600)
		return err
	}

	return nil
}

// Lock the PID file and raise panic if the file is already locked.
func (p *PIDFile) Lock() {
	if err := p.TryLock(); err != nil {
		log.Fatal().Err(err).Str("path", p.path).Msg("failed to lock the PID file")
	}

	log.Info().Str("name", p.name).Msg("locked the PID file")
}

// Release the PID file instance.
func (p *PIDFile) Release() {
	switch _, err := os.Stat(p.path); err {
	case nil:
		if err := os.Remove(p.path); err != nil {
			log.Warn().Err(err).Str("path", p.path).Msg("failed to remove the PID file")
		}
	}

	log.Info().Str("name", p.name).Msg("releasing the PID file")
}
