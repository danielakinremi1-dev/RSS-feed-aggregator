// Package cli implements gator's commands.
package cli

import (
	"github.com/danielakinremi1-dev/RSS-feed-aggregator/internal/config"
	"github.com/danielakinremi1-dev/RSS-feed-aggregator/internal/database"
)

// State is what every command runs against.
type State struct {
	Cfg *config.Config
	DB  *database.Queries
}
