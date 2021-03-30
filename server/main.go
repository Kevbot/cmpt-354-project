package main

import (
	"draco/models"
	"draco/models/postgresql"
	"flag"
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/lib/pq"
)

type application struct {
	jwtSigningKey string
	players       interface {
		Insert(username string, password string, name string) error
		Authenticate(username string, password string) (string, error)
		Get(username string) (*models.Player, error)
	}
	characters interface {
		Insert(c models.Character) (int, error)
		Get(id int) (*models.Character, error)
		GetAllUserCharacters(username string) (*[]models.Character, error)
	}
}

func main() {
	configFile := flag.String("cfg", "config.yml", "Configuration YAML File")
	flag.Parse()

	cfg := createConfigFromFile(*configFile)

	db := initDBConn(cfg.CreatePostgreSQLDBConnString(false))
	fmt.Println(db)

	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:8080"},
	}))

	if cfg.IsProduction {
		// For production builds, we will want to serve the minified
		// application bundle for frontend codes. This is created using
		// the `npm run build` command from the `/client/` directory of
		// the project.
		e.Static("/", "../client/dist")
	}

	app := &application{
		jwtSigningKey: cfg.JWTSigningKey,
		players:       &postgresql.PlayerModel{DB: db},
		characters:    &postgresql.CharacterModel{DB: db},
	}

	if cfg.IsTest {
		app.runTests(db)
	}

	app.registerRoutes(e)

	e.Logger.Fatal(e.Start(":" + strconv.Itoa(cfg.HTTPServer.Port)))
}
