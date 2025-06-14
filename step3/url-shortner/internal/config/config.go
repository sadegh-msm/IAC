package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port            string
	MongoHost       string
	MongoDatabase   string
	RedisHost       string
	MongoCollection string
}

var AppConfig Config

func LoadConfig() {
	viper.SetDefault("PORT", 80)
	viper.SetDefault("MONGOHOST", "localhost")
	viper.SetDefault("REDISHOST", "localhost")
	viper.SetDefault("MONGODATABASE", "urlshortener")
	viper.SetDefault("MONGOCOLLECTION", "urls")

	viper.BindEnv("PORT")
	viper.BindEnv("MONGOHOST")
	viper.BindEnv("REDISHOST")
	viper.BindEnv("MONGODATABASE")
	viper.BindEnv("MONGOCOLLECTION")

	AppConfig = Config{
		Port:            viper.GetString("PORT"),
		MongoHost:       viper.GetString("MONGOHOST"),
		RedisHost:       viper.GetString("REDISHOST"),
		MongoDatabase:   viper.GetString("MONGODATABASE"),
		MongoCollection: viper.GetString("MONGOCOLLECTION"),
	}
}
