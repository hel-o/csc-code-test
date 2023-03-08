package main

import (
	"context"
	"csc-code-test/internal/apis"
	"csc-code-test/internal/logger"
	"csc-code-test/internal/settings"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	settings.LoadSettings()
	defer logger.Logger.Sync()

	apis.InitializeJobQueueManager()
	apis.JobQueueManagerShared.StartWorker()

	e := echo.New()
	e.Use(middleware.Logger())

	apiJob := e.Group("/api/v1")
	apis.RouteApiJobs(apiJob)

	go func() {
		signalStop := make(chan os.Signal, 1)
		signal.Notify(signalStop, syscall.SIGTERM, syscall.SIGINT)
		<-signalStop

		if err := e.Shutdown(context.Background()); err != nil {
			logger.Logger.Error("err shutdown api", zap.Error(err))
		}
	}()

	if err := e.Start(":5000"); err != nil && err != http.ErrServerClosed {
		logger.Logger.Error("err api", zap.Error(err))
	} else {
		logger.Logger.Info("api stopped")
	}

	apis.JobQueueManagerShared.StopWorker()
}
