package api

import (
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/logger"
	"testing"
)

var a *Api

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	a = New(cfg, log)
	m.Run()
}
