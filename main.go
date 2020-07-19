// 34/51

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/inerts73/tronicscorp/config"
	"github.com/inerts73/tronicscorp/handlers"
	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func main() {
	e := echo.New()
	e.POST("/products", handlers.CreateProducts)
	e.Logger.Infof("Listening on %s:%s", cfg.Host, cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)))
}
