package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"danmakustream/user-service/internal/config"
	membershiphandler "danmakustream/user-service/internal/handler/v1/membership"
	"danmakustream/user-service/internal/server"
	"danmakustream/user-service/internal/svc"
)

func main() {
	configFile := flag.String("f", "etc/config.yaml", "configuration file")
	flag.Parse()
	c, err := config.Load(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		log.Fatal(err)
	}
	membershiphandler.StartExpirationWorker(ctx)
	r := server.Router(ctx)
	server := &http.Server{Addr: fmt.Sprintf("%s:%d", c.Host, c.Port), Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdown)
}
