package main

import (
	"log"
	"net/http"

	"github.com/YPS-Database/yps-db-backend/yps"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type PingResponse struct {
	OK bool `json:"ok"`
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, PingResponse{OK: true})
}

func main() {
	// loading config
	config, err := yps.LoadConfig()
	if err != nil {
		log.Fatal("LoadConfig failed:", err)
	}

	// upgrading db
	m, err := migrate.New("file://./migrations", config.DatabaseUrl)
	if err != nil {
		log.Fatal("DB Migrations loading failed:", err)
	}
	err = m.Up()
	if err != nil {
		log.Fatal("DB Migration failure:", err)
	}

	// api router
	router := gin.Default()
	router.GET("/", ping)

	router.Run(config.Address)
}
