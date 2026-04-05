package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/vasatavit-hub/Gator/internal/config"
	"github.com/vasatavit-hub/Gator/internal/database"
)

func main() {

	fmt.Print("Reading config...")
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(" OK")

	dbURL := cfg.Db_url
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dbQueries := database.New(db)

	stateVar := state{
		cfg: &cfg,
		db:  dbQueries,
	}

	commandsVar := &commands{
		commandMap: make(map[string]func(*state, command) error),
	}
	commandsVar.register("login", handlerLogin)
	commandsVar.register("register", handlerRegister)
	commandsVar.register("reset", handlerReset)
	commandsVar.register("users", handlerGetUsers)
	commandsVar.register("agg", handlerAgg)
	commandsVar.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commandsVar.register("feeds", handlerListFeeds)
	commandsVar.register("follow", middlewareLoggedIn(handlerFollow))
	commandsVar.register("following", middlewareLoggedIn(handlerListFeedFollows))
	commandsVar.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	commandsVar.register("scrape", handlerScrapeFeeds)
	commandsVar.register("browse", middlewareLoggedIn(handlerBrowsePosts))

	commandLineArguments := os.Args
	if len(commandLineArguments) < 2 {
		log.Fatal("No command provided")
		return
	}

	commandName := commandLineArguments[1]
	commandArguments := commandLineArguments[2:]

	err = (*commandsVar).run(&stateVar, command{
		name:    commandName,
		command: commandArguments,
	})
	if err != nil {
		log.Fatal(err)
	}

}
