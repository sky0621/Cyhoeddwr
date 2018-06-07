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

	envProjectID := os.Getenv("PROJECT_ID")
	envCredentialsPath := os.Getenv("CREDENTIALS_PATH")
	envAPIKey := os.Getenv("API_KEY")

	e := echo.New()
	e.GET("/:text", func(c echo.Context) error {
		uuid := uuid.NewV4().String()
		logger.Info("START", zap.String("UUID", uuid))

		if envAPIKey != "" {
			apiKeys := c.Request().Header["API_KEY"]
			if apiKeys == nil || len(apiKeys) < 1 {
				logger.Error("No Header[API_KEY]")
				return response(c, http.StatusUnauthorized)
			}
			if envAPIKey != apiKeys[0] {
				logger.Error("Header[API_KEY] is not match", zap.String("HeaderAPIKey", apiKeys[0]))
				return response(c, http.StatusUnauthorized)
			}
		}

		text := c.Param("text")
		logger.Info("/msg/:text", zap.String("text", text))

		m := &Message{
			UUID: uuid,
			Msg:  text,
			Ts:   time.Now().UnixNano(),
		}

		req, err := json.Marshal(m)
		if err != nil {
			logger.Error("Error@json.Marshal", zap.String("ERROR", err.Error()))
			return response(c, http.StatusInternalServerError)
		}

		fmt.Println(string(req))

		ctx := context.Background()
		var cli *pubsub.Client
		if envCredentialsPath == "" {
			cli, err = pubsub.NewClient(ctx, envProjectID)
		} else {
			cli, err = pubsub.NewClient(ctx, envProjectID, option.WithCredentialsFile(envCredentialsPath))
		}
		if err != nil {
			logger.Error("Error@json.Marshal", zap.String("ERROR", err.Error()))
			logger.Error(err.Error())
			return response(c, http.StatusInternalServerError)
		}
		defer func() {
			if cli != nil {
				cli.Close()
			}
		}()

		tpc := cli.Topic("my-topic-a")
		res := tpc.Publish(ctx, &pubsub.Message{
			Attributes: map[string]string{
				"UUID": uuid,
			},
			Data: []byte(text),
		})
		serverID, err := res.Get(ctx)
		if err != nil {
			return response(c, http.StatusInternalServerError)
		}

		// gcloud alpha pubsub subscriptions pull my-subscription-a

		logger.Info("END", zap.String("UUID", uuid), zap.String("serverID", serverID))
		return response(c, http.StatusOK)
	})
	e.Logger.Fatal(e.Start(":14080"))
}

// Message ...
type Message struct {
	UUID string `json:"uuid"`
	Msg  string `json:"msg"`
	Ts   int64  `json:"ts"`
}

func response(c echo.Context, statusCode int) error {
	return c.String(statusCode, http.StatusText(statusCode))
}
