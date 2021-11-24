package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/http2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Post struct {
	Id       int        `json:"id"`
	Title    string     `json:"title"`
	Comments []*Comment `json:"comments"`
}

type Comment struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
	PostId  int    `json:"post_id"`
	Status  string `json:"status"`
}

type event struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

var posts = make(map[int]*Post)

func handleEvent(e *event) error {
	if e.Type == "PostCreated" {

		post := Post{}
		x, _ := json.Marshal(e.Data)

		_ = json.Unmarshal(x, &post)

		posts[post.Id] = &post
	}

	if e.Type == "CommentCreated" {
		comment := Comment{}
		x, _ := json.Marshal(e.Data)

		_ = json.Unmarshal(x, &comment)

		p := posts[comment.PostId]

		if p == nil {
			return errors.New("post not found")
		}

		p.Comments = append(p.Comments, &comment)
	}

	if e.Type == "CommentUpdated" {
		comment := Comment{}
		x, _ := json.Marshal(e.Data)

		_ = json.Unmarshal(x, &comment)

		p := posts[comment.PostId]

		if p == nil {
			return errors.New("post not found")
		}

		for i, v := range p.Comments {
			if v.Id == comment.Id {
				p.Comments[i] = &comment
				break
			}
		}
	}

	return nil
}

func fetchEvents() {

	go func() {
		resp, err := http.Get("http://event-bus-srv:4005/events")
		if err != nil {
			log.Println(err)
			return
		}

		body, _ := ioutil.ReadAll(resp.Body)

		var events []event

		_ = json.Unmarshal(body, &events)

		for _, event := range events {
			fmt.Println("Handling Event: ", event.Type)
			_ = handleEvent(&event)
		}
	}()

}

func main() {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))
	e.Use(middleware.CORS())

	e.GET("/posts", func(ctx echo.Context) error {

		return ctx.JSON(http.StatusOK, posts)
	})

	e.POST("/events", func(ctx echo.Context) error {

		e := new(event)
		if err := ctx.Bind(e); err != nil {
			log.Fatal("Error: ", err)
			return err
		}

		err := handleEvent(e)
		if err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, "OK")
	})

	fetchEvents()

	startServer(e)
}

func startServer(e *echo.Echo) {

	port := os.Getenv("PORT")
	if port == "" {
		port = "4002"
	}

	s := &http2.Server{
		MaxConcurrentStreams: 250,
		MaxReadFrameSize:     1048576,
		IdleTimeout:          10 * time.Second,
	}
	e.Logger.Fatal(e.StartH2CServer(":"+port, s))
}
