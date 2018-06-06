package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"google.golang.org/api/option"

	"golang.org/x/net/context"

	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"

	"cloud.google.com/go/pubsub"
)

// 試作版
func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	projectID := os.Getenv("PROJECT_ID")
	credentialsPath := os.Getenv("CREDENTIALS_PATH")

	e := echo.New()
	e.GET("/:text", func(c echo.Context) error {
		text := c.Param("text")
		logger.Info("/msg/:text", zap.String("text", text))

		uuid := uuid.NewV4()

		m := &Message{
			UUID: uuid.String(),
			Msg:  text,
			Ts:   time.Now().UnixNano(),
		}

		req, err := json.Marshal(m)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		fmt.Println(string(req))

		ctx := context.Background()
		var cli *pubsub.Client
		if credentialsPath == "" {
			cli, err = pubsub.NewClient(ctx, projectID)
		} else {
			cli, err = pubsub.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsPath))
		}
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		defer func() {
			if cli != nil {
				cli.Close()
			}
		}()

		tpc := cli.Topic("my-topic-a")
		res := tpc.Publish(ctx, &pubsub.Message{
			Attributes: map[string]string{
				"UUID": uuid.String(),
			},
			Data: []byte(text),
		})
		serverID, err := res.Get(ctx)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// gcloud alpha pubsub subscriptions pull my-subscription-a

		return c.String(http.StatusOK, serverID)
	})
	e.Logger.Fatal(e.Start(":14080"))
}

// Message ...
type Message struct {
	UUID string `json:"uuid"`
	Msg  string `json:"msg"`
	Ts   int64  `json:"ts"`
}
