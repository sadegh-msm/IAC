package api

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	Ctx         = context.Background()
	RedisClient *redis.Client
	MongoCol    *mongo.Collection
)

func SetupRouter() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/healthz", health)

	e.POST("/shorten", shortenURL)
	e.GET("/:hsh", resolveURL)
	e.DELETE("/:hsh", deleteURL)

	return e
}

func health(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

func shortenURL(c echo.Context) error {
	type Request struct {
		URL    string `json:"url"`
		Expire int    `json:"expire"` // in minutes
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}

	id, err := generateID()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "could not generate ID"})
	}

	expireTime := time.Now().Add(time.Duration(req.Expire) * time.Minute)
	url := URL{
		ID:        id,
		Original:  req.URL,
		CreatedAt: time.Now(),
		ExpireAt:  expireTime,
	}

	_, err = MongoCol.InsertOne(Ctx, url)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "DB insert failed"})
	}

	RedisClient.Set(Ctx, "short:"+id, req.URL, time.Duration(req.Expire)*time.Minute)

	return c.JSON(http.StatusOK, echo.Map{"short_url": c.Scheme() + "://" + c.Request().Host + "/" + id})
}

func resolveURL(c echo.Context) error {
	id := c.Param("hsh")
	key := "short:" + id

	original, err := RedisClient.Get(Ctx, key).Result()
	if err == redis.Nil {
		var result URL
		err := MongoCol.FindOne(Ctx, bson.M{"_id": id}).Decode(&result)
		if err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "URL not found"})
		}
		RedisClient.Set(Ctx, key, result.Original, time.Until(result.ExpireAt))
		original = result.Original
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Redis error"})
	}

	return c.Redirect(http.StatusMovedPermanently, original)
}

func deleteURL(c echo.Context) error {
	id := c.Param("hsh")
	_, err := MongoCol.DeleteOne(Ctx, bson.M{"_id": id})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "DB delete failed"})
	}
	RedisClient.Del(Ctx, "short:"+id)
	return c.JSON(http.StatusOK, echo.Map{"message": "URL deleted"})
}
