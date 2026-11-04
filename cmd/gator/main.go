package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/config"
	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

func main() {

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("read config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	dbQueries := database.New(db)

	programState := &state{cfg: &cfg, db: dbQueries}
	mainCommands := commands{registeredCommands: map[string]func(*state, command) error{}}
	mainCommands.register("login", handlerLogin)
	mainCommands.register("register", handlerRegister)
	mainCommands.register("reset", handlerReset)
	mainCommands.register("users", handlerUsers)
	mainCommands.register("agg", handlerAgg)
	mainCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	mainCommands.register("feeds", handlerFeeds)
	mainCommands.register("follow", middlewareLoggedIn(handlerFollow))
	mainCommands.register("following", middlewareLoggedIn(handlerFollowing))
	mainCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	mainCommands.register("browse", middlewareLoggedIn(handlerBrowse))

	if len(os.Args) < 2 {
		log.Fatal("Usage: cli <command> [args...]")
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	err = mainCommands.run(programState, command{Name: cmdName, Args: cmdArgs})
	if err != nil {
		fmt.Fprintln(os.Stderr, "gator:", err)
		os.Exit(1)
	}

}
