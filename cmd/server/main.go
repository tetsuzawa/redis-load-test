package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/redis/go-redis/v9"
	"github.com/tetesuzawa/redisloadtest"
	"net/http"
	"os"
)

var (
	rdb *redis.ClusterClient
)

func main() {
	// Redisのクライアントを生成
	host := flag.String("host", "127.0.0.1", "host")
	port := flag.Uint("port", 6379, "host")
	flag.Parse()

	rdb = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    []string{fmt.Sprintf("%s:%d", *host, *port)},
		Password: "", // no password set
	})
	defer rdb.Close()
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Errorf("failed to ping redis: %v", err)
		os.Exit(1)
	}

	e := echo.New()
	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// ルートパスに対するハンドラーを定義
	e.GET("/read", read)

	e.GET("/write", write)

	// 8080ポートでサーバーを起動
	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal(err.Error())
	}
}

func read(c echo.Context) error {
	key, err := redisloadtest.GenerateKey(9)
	if err != nil {
		c.Logger().Errorf("failed to generate key: %v", err)
		return c.String(http.StatusInternalServerError, "failed to generate key")
	}
	found, err := rdb.Exists(c.Request().Context(), key).Result()
	if err != nil {
		c.Logger().Errorf("failed to set key: %v", err)
		return c.String(http.StatusInternalServerError, "failed to set key")
	}
	if found == 0 {
		return c.String(http.StatusOK, fmt.Sprintf("not found: %s", key))
	} else {
		return c.String(http.StatusOK, fmt.Sprintf("found: %s", key))
	}
}

func write(c echo.Context) error {
	key, err := redisloadtest.GenerateKey(9)
	if err != nil {
		c.Logger().Errorf("failed to generate key: %v", err)
		return c.String(http.StatusInternalServerError, "failed to generate key")
	}
	err = rdb.SetEx(c.Request().Context(), key, "", redisloadtest.ExpiresSeconds).Err()
	if err != nil {
		c.Logger().Errorf("failed to set key: %v", err)
		return c.String(http.StatusInternalServerError, "failed to set key")
	}

	return c.String(http.StatusOK, fmt.Sprintf("write successfully: %s", key))
}
