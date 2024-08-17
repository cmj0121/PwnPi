// Run the pwn tool with the given configuration.
package pwnpi

import (
	"context"
)

// The Pwn instance that control the PwnPi CLI and how it behaves.
type Pwn struct{}

// Run the PwnPi CLI based on the current configuration
func (p *Pwn) Run(ctx context.Context) error {
	return nil
}
