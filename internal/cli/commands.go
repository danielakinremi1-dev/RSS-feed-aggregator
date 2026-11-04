package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

type command struct {
	Name string
	Args []string
}

// handlerFunc is the signature every command implements. ctx is cancelled
// when the user interrupts the program.
type handlerFunc func(ctx context.Context, s *State, cmd command) error

type commandSpec struct {
	name    string
	usage   string // arguments, e.g. "<name> <url>"
	summary string
	run     handlerFunc
}

type commands struct {
	specs  []commandSpec // registration order, for help output
	byName map[string]commandSpec
}

// Run dispatches args (command name first) to the matching command.
func Run(ctx context.Context, s *State, args []string) error {
	if len(args) < 1 {
		return errors.New("no command given (run 'gator help')")
	}
	return newCommands().run(ctx, s, command{Name: args[0], Args: args[1:]})
}

// IsHelp reports whether args ask for help. Callers check this before loading
// config so help works on a machine that isn't set up yet.
func IsHelp(args []string) bool {
	if len(args) == 0 {
		return true
	}
	switch args[0] {
	case "help", "-h", "--help":
		return true
	}
	return false
}

// PrintHelp writes the list of commands to out.
func PrintHelp(out io.Writer) { newCommands().printHelp(out) }

func newCommands() *commands {
	c := &commands{byName: map[string]commandSpec{}}
	c.register("register", "<name>", "Create a user and log in as them", handlerRegister)
	c.register("login", "<name>", "Switch to an existing user", handlerLogin)
	c.register("users", "", "List all users", handlerUsers)
	c.register("addfeed", "<name> <url>", "Add a feed and follow it", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", "", "List all feeds and their fetch status", handlerFeeds)
	c.register("follow", "<url>", "Follow an existing feed", middlewareLoggedIn(handlerFollow))
	c.register("following", "", "List feeds you follow", middlewareLoggedIn(handlerFollowing))
	c.register("unfollow", "<url>", "Stop following a feed", middlewareLoggedIn(handlerUnfollow))
	c.register("browse", "[limit]", "Show the newest posts from feeds you follow (default 10)", middlewareLoggedIn(handlerBrowse))
	c.register("agg", "[--workers N] <interval>", "Fetch feeds every interval, N at a time (default 4)", handlerAgg)
	c.register("reset", "--force", "Delete ALL users, feeds, follows and posts", handlerReset)
	c.register("help", "", "Show this help", func(context.Context, *State, command) error {
		c.printHelp(os.Stdout)
		return nil
	})
	return c
}

func (c *commands) register(name, usage, summary string, f handlerFunc) {
	spec := commandSpec{name: name, usage: usage, summary: summary, run: f}
	c.specs = append(c.specs, spec)
	c.byName[name] = spec
}

func (c *commands) run(ctx context.Context, s *State, cmd command) error {
	spec, ok := c.byName[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command %q (run 'gator help')", cmd.Name)
	}
	return spec.run(ctx, s, cmd)
}

func (c *commands) printHelp(out io.Writer) {
	fmt.Fprintln(out, "Usage: gator <command> [args...]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Commands:")
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for _, spec := range c.specs {
		fmt.Fprintf(w, "  %s\t%s\n", strings.TrimSpace(spec.name+" "+spec.usage), spec.summary)
	}
	w.Flush()
}
