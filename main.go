package main

import (
	"log"

	"github.com/YPS-Database/yps-db-backend/yps"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// loading config
	config, err := yps.LoadConfig()
	if err != nil {
		log.Fatal("LoadConfig failed:", err)
	}

	// upgrading db
	m, err := migrate.New("file://"+config.DatabaseMigrationsPath, config.DatabaseUrl)
	if err != nil {
		log.Fatal("DB Migrations loading failed:", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal("DB Migration failure:", err)
	}
	m.Close()

	// api router
	router := yps.GetRouter()
	router.Run(config.Address)
}
