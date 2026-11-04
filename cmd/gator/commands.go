package main

import (
	"context"
	"fmt"
)

type command struct {
	Name string
	Args []string
}

// handlerFunc is the signature every command implements. ctx is cancelled
// when the user interrupts the program.
type handlerFunc func(ctx context.Context, s *state, cmd command) error

type commands struct {
	registeredCommands map[string]handlerFunc
}

func (c *commands) run(ctx context.Context, s *state, cmd command) error {
	cmdFunc, ok := c.registeredCommands[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command %q", cmd.Name)
	}

	return cmdFunc(ctx, s, cmd)
}

func (c *commands) register(name string, f handlerFunc) {
	c.registeredCommands[name] = f
}
