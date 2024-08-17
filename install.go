// Install the pwnpi package into the RPi.
package pwnpi

import (
	"context"
	_ "embed"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:embed assets/sytem/pwnpi.service
var systemdService string

// The install command is used to install the PwnPi CLI into the Raspberry Pi.
type Install struct {
	Prefix string `arg:"" default:"/usr/local/sbin" help:"The prefix path to install the PwnPi CLI."`
}

// Run the install command to install the PwnPi CLI into the Raspberry Pi.
func (i *Install) Run(ctx context.Context) error {
	// save the systemd service file into the target path
	path := "/etc/systemd/system/pwnpi.service"
	log.Trace().Str("path", path).Msg("writing the systemd service file")

	err := os.WriteFile(path, []byte(systemdService), 0644)
	if err != nil {
		log.Warn().Err(err).Str("path", path).Msg("failed to write the systemd service file")
		return err
	}

	return i.restart()
}

// Restart the PwnPi service after the installation.
func (i *Install) restart() error {
	commands := [][]string{
		{"systemctl", "daemon-reload"},
		{"systemctl", "enable", "pwnpi"},
		{"systemctl", "restart", "pwnpi"},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		log.Info().Str("shell", strings.Join(args, " ")).Msg("running the command")
		err := cmd.Run()

		if err != nil {
			log.Warn().Err(err).Str("shell", strings.Join(args, " ")).Msg("failed to run the command")
			return err
		}
	}

	log.Debug().Msg("successfully installed the PwnPi CLI into the Raspberry Pi")
	return nil
}
