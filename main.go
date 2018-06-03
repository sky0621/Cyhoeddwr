package main

import (
	"net/http"

	"github.com/labstack/echo"
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

		// FIXME send text to CloudPubSub

		return c.String(http.StatusOK, text)
	})
	e.Logger.Fatal(e.Start(":14080"))
}
