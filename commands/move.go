package commands

import (
	"github.com/pspiagicw/groove/config"
	"github.com/pspiagicw/groove/database"
)

func Move(configPath string) error {
	config := config.ConfigProvider(configPath)

	db := database.NewDB(config.Database)

	files := db.QueryFiles()
}
