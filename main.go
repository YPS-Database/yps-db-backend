package main

import (
	"log"
	"strings"

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

	// setup auth
	err = yps.SetupAuth(config.PasetoKey, config.AdminPass, config.SuperuserPass)
	if err != nil {
		log.Fatal("SetupAuth failed:", err)
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

	// open db
	err = yps.OpenDatabase(config.DatabaseUrl)
	if err != nil {
		log.Fatal("Opening db failed:", err)
	}

	// set browse by fields
	err = yps.UpdateBrowseByFields()
	if err != nil {
		log.Fatal("Setting browse by fields failed:", err)
	}

	// api router
	router := yps.GetRouter(nil, strings.Split(config.CorsAllowedFrom, " "))
	router.Run(config.Address)
}
