package api

import (
	"github.com/aibotsoft/micro/config"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type Api struct {
	cfg    *config.Config
	log    *zap.SugaredLogger
	client *resty.Client
}

func New(cfg *config.Config, log *zap.SugaredLogger) *Api {
	client := resty.New()
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	client.SetHeader("User-Agent", `Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1`)
	client.SetHeader("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`)
	client.SetHeader("Accept-Language", `en-US;q=0.8,en;q=0.7`)
	client.SetHeader("Connection", `keep-alive`)
	client.SetHeader("Cache-Control", `no-cache`)
	client.SetHeader("Pragma", `no-cache`)
	client.SetDebug(cfg.Service.Debug)
	return &Api{cfg: cfg, log: log, client: client}
}
