package main

import (
	"log"
	"url-shortner/internal/api"
	"url-shortner/internal/config"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	e := api.SetupRouter()

	config.LoadConfig()

	api.RedisClient = redis.NewClient(&redis.Options{
		Addr: config.AppConfig.RedisHost,
	})

	client, err := mongo.Connect(api.Ctx, options.Client().ApplyURI("mongodb://"+config.AppConfig.MongoHost))
	if err != nil {
		log.Fatal(err)
	}
	api.MongoCol = client.Database(config.AppConfig.MongoDatabase).Collection(config.AppConfig.MongoCollection)

	e.Logger.Fatal(e.Start("0.0.0.0:" + config.AppConfig.Port))
}
