package collector

import (
	"context"
	"github.com/aibotsoft/daf-service/pkg/api"
	"github.com/aibotsoft/daf-service/pkg/store"
	"github.com/aibotsoft/daf-service/services/auth"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
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
	conf := config_client.New(cfg, log)
	serviceApi := api.New(cfg, log)
	au := auth.New(cfg, log, sto, serviceApi, conf)
	c = New(cfg, log, serviceApi, sto, au)
	m.Run()
	sto.Close()
}

func TestCollector_CollectRound(t *testing.T) {
	err := c.CollectRound()
	assert.NoError(t, err)
}

func TestCollector_Markets(t *testing.T) {
	var events []api.Event
	for _, state := range eventStates {
		sports, err := c.api.GetSports(context.Background(), c.auth.Base(), state)
		c.errLog(err)
		for _, sport := range sports {
			leagues, err := c.api.GetLeagues(context.Background(), c.auth.Base(), sport.EventState, sport.Id)
			c.errLog(err)
			for _, league := range leagues {
				event, err := c.api.GetEvents(context.Background(), c.auth.Base(), league.EventState, league.SportId, league.Id)
				c.errLog(err)
				events = append(events, event...)
			}
		}
	}
	c.log.Infow("begin save events", "count", len(events))
	err := c.store.SaveEvents(context.Background(), events)
	if err != nil {

	}
	c.errLog(err)

	//leagues, err:=c.Leagues(context.Background(), "today", 1)
	//assert.NoError(t, err)
	//t.Log(leagues)
	//got, err := c.api.GetMarkets(context.Background(),c.auth.Base(), "earlyMarket", 1, 71779, 36003515)
	////t.Log(got)
	//assert.NoError(t, err)
	//err = c.store.SaveLines(context.Background(), got)
	//assert.NoError(t, err)
	//for _, ticket := range got {
	//	c.log.Infow("", "", ticket.Points)
	//}
	//for _, event := range events {
	//	got, err := c.Markets(context.Background(), event)
	//	if assert.NoError(t, err) {
	//		assert.NotEmpty(t, got)
	//		t.Log(got)
	//		err = c.store.SaveLines(context.Background(), got)
	//		assert.NoError(t, err)
	//	}
	//}

}
