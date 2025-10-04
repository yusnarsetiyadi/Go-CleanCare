package main

import (
	"context"
	"iss_cleancare/internal/config"
	"iss_cleancare/internal/factory"
	httpiss_cleancare "iss_cleancare/internal/http"
	middlewareEcho "iss_cleancare/internal/middleware"
	db "iss_cleancare/pkg/database"
	"iss_cleancare/pkg/log"
	"iss_cleancare/pkg/ngrok"
	"iss_cleancare/pkg/ws"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// @title iss_cleancare
// @version 1.0.0
// @description This is a doc for iss_cleancare

func main() {
	config.Init()

	log.Init()

	db.Init()

	e := echo.New()

	f := factory.NewFactory()

	middlewareEcho.Init(e, f.DbRedis)

	httpiss_cleancare.Init(e, f)

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ws.InitCentrifugal(ctx, e, f)

	go func() {
		runNgrok := true
		addr := ""
		if runNgrok {
			listener := ngrok.Run()
			e.Listener = listener
			addr = "/"
		} else {
			addr = ":" + config.Get().App.Port
		}
		err := e.Start(addr)
		if err != nil {
			if err != http.ErrServerClosed {
				logrus.Fatal(err)
			}
		}
	}()

	<-ch

	logrus.Println("Shutting down server...")
	cancel()

	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	e.Shutdown(ctx2)
	logrus.Println("Server gracefully stopped")
}
