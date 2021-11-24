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
	"strconv"
	"time"
)

type Comment struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
	PostId  int    `json:"post_id"`
	Status  string `json:"status"`
}

type event struct {
	Type string
	Data map[string]interface{}
}

var commentsByPostId = make(map[int][]*Comment)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))
	e.Use(middleware.CORS())

	e.GET("/posts/:id/comments", func(ctx echo.Context) error {
		postId, _ := strconv.Atoi(ctx.Param("id"))

		c := commentsByPostId[postId]

		if c == nil {
			//TODO(Read on returning nil or [])
			return ctx.JSON(http.StatusOK, []string{})
		}

		return ctx.JSON(http.StatusOK, c)

	})

	e.POST("/posts/:id/comments", func(ctx echo.Context) error {
		id := rand.Int()
		postId := ctx.Param("id")

		if postId == "undefined" {
			return ctx.JSON(http.StatusUnprocessableEntity, "Post Id is not specified")
		}

		c := new(Comment)
		if err := ctx.Bind(c); err != nil {
			return err
		}

		c.Id = id
		c.PostId, _ = strconv.Atoi(postId)
		c.Status = "pending"

		comments := commentsByPostId[c.PostId]
		comments = append(comments, c)

		commentsByPostId[c.PostId] = comments

		go func() {

			x, _ := json.Marshal(c)
			inInterface := make(map[string]interface{})

			_ = json.Unmarshal(x, &inInterface)

			postBody, _ := json.Marshal(event{
				Type: "CommentCreated",
				Data: inInterface,
			})

			_, _ = http.Post("http://event-bus-srv:4005/events", "application/json", bytes.NewBuffer(postBody))
		}()

		return ctx.JSON(http.StatusCreated, c)
	})

	e.POST("/events", func(ctx echo.Context) error {
		e := new(event)
		if err := ctx.Bind(e); err != nil {
			log.Fatal("Error: ", err)
			return err
		}

		fmt.Println("Received Event", e.Type)

		if e.Type == "CommentModerated" {
			comment := Comment{}
			x, _ := json.Marshal(e.Data)

			_ = json.Unmarshal(x, &comment)

			p := commentsByPostId[comment.PostId]

			if p == nil {
				return ctx.JSON(http.StatusNotFound, "Post not found")
			}

			for i, v := range p {
				if v.Id == comment.Id {
					p[i] = &comment
					break
				}
			}

			go func() {

				x, _ := json.Marshal(comment)
				inInterface := make(map[string]interface{})

				_ = json.Unmarshal(x, &inInterface)

				postBody, _ := json.Marshal(event{
					Type: "CommentUpdated",
					Data: inInterface,
				})

				_, _ = http.Post("http://event-bus-srv:4005/events", "application/json", bytes.NewBuffer(postBody))
			}()
		}

		return ctx.JSON(http.StatusOK, "OK")
	})
	s := &http2.Server{
		MaxConcurrentStreams: 250,
		MaxReadFrameSize:     1048576,
		IdleTimeout:          10 * time.Second,
	}
	e.Logger.Fatal(e.StartH2CServer(":4001", s))
}
