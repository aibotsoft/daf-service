package collector

import (
	"context"
	"github.com/aibotsoft/daf-service/pkg/api"
	"github.com/aibotsoft/daf-service/pkg/store"
	"github.com/aibotsoft/daf-service/services/auth"
	"github.com/aibotsoft/micro/config"
	"go.uber.org/zap"
	"time"
)

type Collector struct {
	cfg   *config.Config
	log   *zap.SugaredLogger
	api   *api.Api
	store *store.Store
	auth  *auth.Auth
}

var eventStates = []string{"today", "earlyMarket"}

const collectMinPeriod = 2 * time.Second

func New(cfg *config.Config, log *zap.SugaredLogger, api *api.Api, store *store.Store, auth *auth.Auth) *Collector {
	return &Collector{cfg: cfg, log: log, api: api, store: store, auth: auth}
}
func (c *Collector) CollectJob() {
	for {
		start := time.Now()
		err := c.CollectRound()
		if err != nil {
			c.log.Error(err)
		}
		c.log.Infow("CollectJob done", "time", time.Since(start))
	}
}

var stopSportId = map[int64]bool{13: true, 55: true}

func (c *Collector) CollectRound() error {
	ctx := context.Background()
	var sports []api.Sport
	var leagues []api.League
	var events []api.Event
	var tickets []api.Ticket
	c.log.Infow("begin collect sports")
	for _, state := range eventStates {
		startRequest := time.Now()
		s, err := c.api.GetSports(ctx, c.auth.Base(), state)
		sports = append(sports, s...)
		c.errLogAndSleep(err, startRequest)
	}
	c.log.Infow("begin save sports", "count", len(sports))
	err := c.store.SaveSports(ctx, sports)
	c.errLog(err)

	c.log.Infow("begin collect leagues", "count", len(sports))
	for _, sport := range sports {
		if stopSportId[sport.Id] {
			continue
		}
		startRequest := time.Now()
		league, err := c.api.GetLeagues(ctx, c.auth.Base(), sport.EventState, sport.Id)
		leagues = append(leagues, league...)
		c.errLogAndSleep(err, startRequest)
	}
	c.log.Infow("begin save leagues", "count", len(leagues))
	err = c.store.SaveLeagues(ctx, leagues)
	c.errLog(err)

	c.log.Infow("begin collect events", "count", len(leagues))
	for _, league := range leagues {
		startRequest := time.Now()
		event, err := c.api.GetEvents(ctx, c.auth.Base(), league.EventState, league.SportId, league.Id)
		c.errLogAndSleep(err, startRequest)
		events = append(events, event...)
	}
	c.log.Infow("begin save events", "count", len(events))
	err = c.store.SaveEvents(ctx, events)
	c.errLog(err)

	c.log.Infow("begin collect markets", "count", len(events))
	for _, event := range events {
		startRequest := time.Now()
		ticket, err := c.api.GetMarkets(ctx, c.auth.Base(), &event)
		c.errLogAndSleep(err, startRequest)
		tickets = append(tickets, ticket...)
	}
	c.log.Infow("begin save lines", "count", len(tickets))
	err = c.store.SaveLines(ctx, tickets)
	c.errLog(err)
	//c.log.Infow("begin collect betList")
	//_, err := c.BetList(ctx)
	//if err != nil {
	//	c.log.Error(err)
	return nil
}
func (c *Collector) errLogAndSleep(err error, start time.Time) {
	if err != nil {
		c.log.Info(err)
	}
	needSleep := collectMinPeriod - time.Since(start)
	//c.log.Debugw("need sleep", "duration", needSleep)
	time.Sleep(needSleep)
}
func (c *Collector) errLog(err error) {
	if err != nil {
		c.log.Info(err)
	}
}
