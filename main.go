package main

import (
	"cleancare/internal/config"
	"cleancare/internal/factory"
	httpcleancare "cleancare/internal/http"
	middlewareEcho "cleancare/internal/middleware"
	db "cleancare/pkg/database"
	"cleancare/pkg/log"
	"cleancare/pkg/ngrok"
	"cleancare/pkg/ws"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// @title cleancare
// @version 1.0.0.
// @description This is a doc for cleancare

func main() {
	config.Init()

	log.Init()

	db.Init()

	e := echo.New()

	f := factory.NewFactory()

	middlewareEcho.Init(e, f.DbRedis)

	httpcleancare.Init(e, f)

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ws.InitCentrifugal(ctx, e, f)

	go func() {
		runNgrok := false
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
