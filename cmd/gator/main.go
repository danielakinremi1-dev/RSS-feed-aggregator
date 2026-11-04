package main

import (
	"database/sql"
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
		log.Fatalf("error reading config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	dbQueries := database.New(db)

	programState := &state{cfg: &cfg, db: dbQueries}
	mainCommands := commands{registeredCommands: map[string]func(*state, command) error{}}
	mainCommands.register("login", handlerLogin)
	mainCommands.register("register", handlerRegister)
	mainCommands.register("reset", handlerReset)
	mainCommands.register("users", handlerUsers)
	mainCommands.register("agg", handlerAgg)
	mainCommands.register("addfeed", midddlewareLoggedIn(handlerAddFeed))
	mainCommands.register("feeds", handlerFeeds)
	mainCommands.register("follow", midddlewareLoggedIn(handlerFollow))
	mainCommands.register("following", midddlewareLoggedIn(handlerFollowing))
	mainCommands.register("unfollow", midddlewareLoggedIn(handlerUnfollow))
	mainCommands.register("browse", midddlewareLoggedIn(handlerBrowse))

	if len(os.Args) < 2 {
		log.Fatal("Usage: cli <command> [args...]")
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	err = mainCommands.run(programState, command{Name: cmdName, Args: cmdArgs})
	if err != nil {
		log.Fatal(err)
	}

}
