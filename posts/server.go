package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/http2"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type Post struct {
	Id    int    `json:"id"`
	Title string `json:"title" validate:"required"`
}

type event struct {
	Type string
	Data interface{}
}

var posts = make(map[int]*Post)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))
	e.Use(middleware.CORS())

	e.GET("/posts", func(c echo.Context) error {
		return c.JSON(http.StatusOK, posts)
	})

	e.POST("/posts/create", func(c echo.Context) error {
		id := rand.Int()

		p := new(Post)
		if err := c.Bind(p); err != nil {
			return err
		}

		if p.Title == "" {
			return c.JSON(http.StatusUnprocessableEntity, "Title is required")
		}

		p.Id = id
		posts[id] = p

		go func() {
			x, _ := json.Marshal(p)
			inInterface := new(interface{})

			_ = json.Unmarshal(x, &inInterface)

			postBody, _ := json.Marshal(event{
				Type: "PostCreated",
				Data: inInterface,
			})

			_, _ = http.Post("http://event-bus-srv:4005/events", "application/json", bytes.NewBuffer(postBody))
		}()

		return c.JSON(http.StatusCreated, posts[id])
	})

	e.POST("/events", func(ctx echo.Context) error {

		e := new(event)
		if err := ctx.Bind(e); err != nil {
			log.Fatal("Error: ", err)
			return err
		}

		fmt.Println("Received Event", e.Type)

		return ctx.JSON(http.StatusOK, "OK")
	})

	s := &http2.Server{
		MaxConcurrentStreams: 250,
		MaxReadFrameSize:     1048576,
		IdleTimeout:          10 * time.Second,
	}
	e.Logger.Fatal(e.StartH2CServer(":4000", s))
}
