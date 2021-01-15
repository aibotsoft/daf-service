package main

import (
	"fmt"
	"github.com/aibotsoft/daf-service/pkg/store"
	"github.com/aibotsoft/daf-service/services/auth"
	"github.com/aibotsoft/daf-service/services/collector"
	"github.com/aibotsoft/daf-service/services/handler"
	"github.com/aibotsoft/daf-service/services/server"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.New()
	log := logger.New()
	log.Infow("Begin service", "config", cfg.Service)
	db := sqlserver.MustConnectX(cfg)
	sto := store.NewStore(cfg, log, db)
	conf := config_client.New(cfg, log)

	au := auth.New(cfg, log, sto, conf)
	go au.AuthJob()
	h := handler.NewHandler(cfg, log, sto, au, conf)
	go h.BalanceJob()
	go h.BetListJob()

	s := server.NewServer(cfg, log, h)
	// Инициализируем Close
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	c := collector.New(cfg, log, sto)
	go c.CollectJob()

	go func() { errc <- s.Serve() }()
	defer func() { s.Close() }()
	log.Info("exit: ", <-errc)
}
