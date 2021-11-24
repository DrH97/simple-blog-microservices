package main

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/http2"
	"log"
	"net/http"
	"time"
)

type event struct {
	Type string
	Data map[string]interface{}
}

var events []event

func main() {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))
	e.Use(middleware.CORS())

	e.POST("/events", func(ctx echo.Context) error {

		e := new(event)
		if err := ctx.Bind(e); err != nil {
			log.Fatal("Error: ", err)
			return err
		}

		events = append(events, *e)

		postBody, _ := json.Marshal(e)

		go func() {
			_, _ = http.Post("http://posts-srv:4000/events", "application/json", bytes.NewBuffer(postBody))
			_, _ = http.Post("http://comments-srv:4001/events", "application/json", bytes.NewBuffer(postBody))
			_, _ = http.Post("http://query-srv:4002/events", "application/json", bytes.NewBuffer(postBody))
			_, _ = http.Post("http://moderation-srv:4003/events", "application/json", bytes.NewBuffer(postBody))
		}()

		return ctx.JSON(http.StatusOK, "OK")
	})


	e.GET("/events", func(ctx echo.Context) error {
		return ctx.JSON(http.StatusOK, events)
	})

	s := &http2.Server{
		MaxConcurrentStreams: 250,
		MaxReadFrameSize:     1048576,
		IdleTimeout:          10 * time.Second,
	}
	e.Logger.Fatal(e.StartH2CServer(":4005", s))
}
