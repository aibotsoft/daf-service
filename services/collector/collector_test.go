package collector

import (
	"context"
	"github.com/aibotsoft/daf-service/pkg/store"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/stretchr/testify/assert"
	"testing"
)

var c *Collector

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	db := sqlserver.MustConnectX(cfg)
	sto := store.NewStore(cfg, log, db)
	c = New(cfg, log, sto)
	m.Run()
	sto.Close()
}

func TestCollector_CollectRound(t *testing.T) {
	err := c.CollectRound(0)
	assert.NoError(t, err)
}

func TestCollector_CollectMarkets(t *testing.T) {
	err := c.CollectMarketsRound()
	assert.NoError(t, err)
}

func TestCollector_CollectMoreEvents(t *testing.T) {
	err := c.CollectMoreEvents(context.Background(), 43, nil)
	assert.NoError(t, err)
}

func TestCollector_CollectJob(t *testing.T) {
	c.CollectJob()
}
