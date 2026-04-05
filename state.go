package main

import (
	"github.com/vasatavit-hub/Gator/internal/config"
	"github.com/vasatavit-hub/Gator/internal/database"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}
