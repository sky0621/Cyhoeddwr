package main

import (
	"bytes"
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

const (
	// TargetPort ...
	TargetPort = ":14080"

	// TargetPath ...
	TargetPath = "/:text"

	// TargetParam ...
	TargetParam = "text"

	// TargetTopic ...
	TargetTopic = "my-topic-a"

	// HeaderAPIKey ...
	HeaderAPIKey = "Cyhoeddwrapikey"
)

// 試作版
func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	env := NewEnv(os.Getenv(EkeyProID), os.Getenv(EkeyCredPath), os.Getenv(EkeyAPIKey))
	logger.Info("ENV", zap.String("env", env.String()))

	e := echo.New()
	e.GET(TargetPath, func(c echo.Context) error {
		uuid := uuid.NewV4().String()
		logger.Info("START", zap.String("UUID", uuid))

		if env.APIKey != "" {
			fmt.Println(c.Request())
			apiKeys := c.Request().Header[HeaderAPIKey]
			logger.Info("Got Header", zap.Strings("apiKeys", apiKeys))
			if apiKeys == nil || len(apiKeys) < 1 {
				logger.Error("No Header[CYHOEDDWR_API_KEY]")
				return response(c, http.StatusUnauthorized)
			}
			if env.APIKey != apiKeys[0] {
				logger.Error("Header[CYHOEDDWR_API_KEY] is not match", zap.String("HeaderAPIKey", apiKeys[0]))
				return response(c, http.StatusUnauthorized)
			}
		}

		text := c.Param(TargetParam)
		logger.Info(TargetPath, zap.String(TargetParam, text))

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
		if env.CredentialsPath == "" {
			cli, err = pubsub.NewClient(ctx, env.ProjectID)
		} else {
			cli, err = pubsub.NewClient(ctx, env.ProjectID, option.WithCredentialsFile(env.CredentialsPath))
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

		tpc := cli.Topic(TargetTopic)
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
	e.Logger.Fatal(e.Start(TargetPort))
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

// ------------------------------------------------------------------
// Env
// ------------------------------------------------------------------
const (
	EkeyProID    = "CYHOEDDWR_PROJECT_ID"
	EkeyCredPath = "CYHOEDDWR_CREDENTIALS_PATH"
	EkeyAPIKey   = "CYHOEDDWR_API_KEY"
)

// Env ...
type Env struct {
	ProjectID       string
	CredentialsPath string
	APIKey          string
}

// NewEnv ...
func NewEnv(envProjectID, envCredentialsPath, envAPIKey string) *Env {
	return &Env{
		ProjectID:       envProjectID,
		CredentialsPath: envCredentialsPath,
		APIKey:          envAPIKey,
	}
}

// String ...
func (e *Env) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	if e == nil {
		buf.WriteString("}")
		return buf.String()
	}
	buf.WriteString(fmt.Sprintf("envProjectID:%s, ", e.ProjectID))
	buf.WriteString(fmt.Sprintf("envCredentialsPath:%s, ", e.CredentialsPath))
	buf.WriteString(fmt.Sprintf("envAPIKey:%s, ", e.APIKey))
	buf.WriteString("}")
	return buf.String()
}
