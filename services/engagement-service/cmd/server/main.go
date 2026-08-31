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

	"danmakustream/engagement-service/internal/app"
	"danmakustream/engagement-service/internal/config"
	"danmakustream/engagement-service/internal/database"
	"danmakustream/engagement-service/internal/handler"
)

var version = "microservice-0.1.0"
var commit = "0000000"
var buildTime = "1970-01-01T00:00:00Z"

func main() {
	configPath := flag.String("f", "etc/config.yaml", "config file")
	flag.Parse()
	c, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	if c.Build.Version == "" || c.Build.Version == "microservice-0.1.0" {
		c.Build.Version = version
	}
	if c.Build.GitSHA == "" || c.Build.GitSHA == "0000000" {
		c.Build.GitSHA = commit
	}
	if c.Build.Time == "" || c.Build.Time == "1970-01-01T00:00:00Z" {
		c.Build.Time = buildTime
	}
	db, err := database.Open(c.Database.DataSource)
	if err != nil {
		log.Fatal(err)
	}
	a := app.New(c, db)
	handler.StartScheduleWorker(db)
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	server := &http.Server{Addr: addr, Handler: a.Router(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		log.Printf("service=%s addr=%s", c.Name, addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("service=%s shutdown_error=%v", c.Name, err)
	}
}
