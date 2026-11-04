package cli

import (
	"context"
	"errors"
	"fmt"
)

type command struct {
	Name string
	Args []string
}

// handlerFunc is the signature every command implements. ctx is cancelled
// when the user interrupts the program.
type handlerFunc func(ctx context.Context, s *State, cmd command) error

// Run dispatches args (command name first) to the matching command.
func Run(ctx context.Context, s *State, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: gator <command> [args...]")
	}
	return newCommands().run(ctx, s, command{Name: args[0], Args: args[1:]})
}

func newCommands() *commands {
	c := &commands{registeredCommands: map[string]handlerFunc{}}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)
	c.register("agg", handlerAgg)
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerFeeds)
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("following", middlewareLoggedIn(handlerFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	c.register("browse", middlewareLoggedIn(handlerBrowse))
	return c
}

type commands struct {
	registeredCommands map[string]handlerFunc
}

func (c *commands) run(ctx context.Context, s *State, cmd command) error {
	cmdFunc, ok := c.registeredCommands[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command %q", cmd.Name)
	}

	return cmdFunc(ctx, s, cmd)
}

func (c *commands) register(name string, f handlerFunc) {
	c.registeredCommands[name] = f
}
