// 35/51

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/inerts73/tronicscorp/config"
	"github.com/inerts73/tronicscorp/handlers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/labstack/gommon/random"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// CorrelationID is a request id unique to the request being made
	CorrelationID = "X-Correlation-ID"
)

var (
	c   *mongo.Client
	db  *mongo.Database
	col *mongo.Collection
	cfg config.Properties
)

func init() {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Configuration cannot be read : %v", err)
	}
	connectURI := fmt.Sprintf("mongodb://%s:%s", cfg.DBHost, cfg.DBPort)
	c, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connectURI))
	if err != nil {
		log.Fatal("Unable tp connect to database: %w", err)
	}
	db = c.Database(cfg.DBName)
	col = db.Collection(cfg.CollectionName)
}

func addCorrelationID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context)error{
		// generate correlation id
		id := c.Request().Header.Get(CorrelationID)
		var newID string
		if id == "" {
			//generate a random number
			newID = random.String(12)
		} else {
			newID = id
		}

		c.Request().Header.Set(CorrelationID, newID)
		c.Response().Header().Set(CorrelationID, newID)
		return next(c)	
	}
}

func main() {
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	middleware.RequestID()
	e.Pre(addCorrelationID)
	h := handlers.ProductHandler{Col: col}
	e.POST("/products", h.CreateProducts, middleware.BodyLimit("1M"))
	e.Logger.Infof("Listening on %s:%s", cfg.Host, cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)))
}
