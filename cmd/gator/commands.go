package main

import (
	"fmt"
)

type command struct {
	Name string
	Args []string
}

type commands struct {
	registeredCommands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	cmdFunc, ok := c.registeredCommands[cmd.Name]
	if !ok {
		return fmt.Errorf("Command not found")
	}

	return cmdFunc(s, cmd)

}

func (c *commands) register(Name string, f func(*state, command) error) {
	c.registeredCommands[Name] = f
}
