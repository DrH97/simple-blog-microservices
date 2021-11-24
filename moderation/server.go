package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/http2"
	"log"
	"net/http"
	"strings"
	"time"
)


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

func main() {
	e := echo.New()
	e.HideBanner = true

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

		fmt.Println("Received Event", e.Type)

		if e.Type == "CommentCreated" {

			comment := Comment{}
			x, _ := json.Marshal(e.Data)

			_ = json.Unmarshal(x, &comment)

			if rejectComment := strings.Contains(comment.Content, "orange"); rejectComment {
				comment.Status = "rejected"
			} else {
				comment.Status = "approved"
			}


			go func() {

				x, _ := json.Marshal(comment)
				inInterface := make(map[string]interface{})

				_ = json.Unmarshal(x, &inInterface)

				postBody, _ := json.Marshal(event{
					Type: "CommentModerated",
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
	e.Logger.Fatal(e.StartH2CServer(":4003", s))
}
