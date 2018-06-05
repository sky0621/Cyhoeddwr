package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
)

// 試作版
func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	e := echo.New()
	e.GET("/msg/:text", func(c echo.Context) error {
		text := c.Param("text")
		logger.Info("/msg/:text", zap.String("text", text))

		uuid, err := uuid.NewV4()
		if err != nil {
			panic(err)
		}

		m := &Message{
			UUID: uuid.String(),
			Msg:  text,
			Ts:   time.Now().UnixNano(),
		}

		req, err := json.Marshal(m)
		if err != nil {
			panic(err)
		}
		// FIXME send text to CloudPubSub
		fmt.Println(req)

		return c.String(http.StatusOK, text)
	})
	e.Logger.Fatal(e.Start(":14080"))
}

type Message struct {
	UUID string `json:"uuid"`
	Msg  string `json:"msg"`
	Ts   int64  `json:"ts"`
}
