package collector

import (
	"context"
	"github.com/aibotsoft/daf-service/pkg/help"
	"github.com/aibotsoft/daf-service/pkg/store"
	api "github.com/aibotsoft/gen/dafapi"
	"github.com/aibotsoft/micro/config"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Collector struct {
	cfg    *config.Config
	log    *zap.SugaredLogger
	client *api.APIClient
	store  *store.Store
}

const (
	collectMinPeriod      = 18 * time.Second
	minEventCountForCheck = 2
	earlyEventStartLimit  = 40 * time.Hour
)

var stopSportId = map[int64]bool{13: true, 55: true, 999: true, 181: true, 182: true, 183: true, 161: true, 10: true, 50: true, 16: true}
var clientLock sync.Mutex
var done = make(chan struct{})
var earlyEventChan = make(chan store.Event, 50000)

func New(cfg *config.Config, log *zap.SugaredLogger, store *store.Store) *Collector {
	clientConfig := api.NewConfiguration()
	clientConfig.Debug = cfg.Service.Debug
	tr := &http.Transport{TLSHandshakeTimeout: 0 * time.Second, IdleConnTimeout: 0 * time.Second, MaxIdleConnsPerHost: 3, MaxConnsPerHost: 3}
	clientConfig.HTTPClient = &http.Client{Transport: tr}
	client := api.NewAPIClient(clientConfig)
	return &Collector{cfg: cfg, log: log, client: client, store: store}
}

func (c *Collector) CollectJob() {
	defer close(done)
	go c.CollectMarketsJob()
	go c.CollectEarlyEvents()
	var roundCount int
	for {
		start := time.Now()
		err := c.CollectRound(roundCount)
		if err != nil {
			c.log.Info(err)
			time.Sleep(time.Minute)
		}
		if time.Since(start) < time.Minute {
			c.log.Infow("collect_job_done_too_fast", "time", time.Since(start))
			time.Sleep(1 * time.Minute)
		} else {
			c.log.Infow("CollectJob_done", "time", time.Since(start))
		}
		roundCount += 1
	}
}

func (c *Collector) CollectRound(roundCount int) error {
	ctx := context.Background()
	clientLock.Lock()
	resp, _, err := c.client.OddsApi.GetSports(ctx, "BF").ContributorRequest(api.ContributorRequest{}).Execute()
	clientLock.Unlock()
	if err != nil || resp.ErrorCode != 0 {
		return errors.Errorf("get_sports_error")
	}
	var sports []store.Sport
	var leagues []store.League
	var teams []store.Team
	var events []store.Event
	var tickets []store.Ticket

	sportList := resp.GetData()
	for i := range sportList {
		sports = append(sports, store.Sport{Id: sportList[i].GameId, Name: sportList[i].Name})
	}
	//c.log.Infow("begin save sports", "count", len(sports))
	err = c.store.SaveSports(ctx, sports)
	c.errLog(err)
	for _, sport := range sportList {
		for _, dateType := range []string{"t", "e"} {
			if stopSportId[sport.GameId] {
				continue
			}
			if eventCont(dateType, sport.M0) < minEventCountForCheck {
				//c.log.Infow("sport has no events","sport", sport.Name, "dt", dateType, "m0", sport.M0, "sportId", sport.GameId)
				continue
			}
			//c.log.Infow("sport", "", sport, "count", eventCont(dateType, sport.M0), "dateType", dateType)
			err := c.CollectEvents(ctx, sport.GameId, dateType, &leagues, &teams, &events, &tickets)
			if err != nil {
				c.log.Info(err)
			}
			if dateType == "e" {
				continue
			}
			if roundCount%2 != 0 {
				//c.log.Infow("это второй луп поэтому не берем more", "roundCount", roundCount)
				continue
			}
			err = c.CollectMoreEvents(ctx, sport.GameId, &tickets)
			if err != nil {
				c.log.Info(err)
			}
		}
	}
	c.log.Infow("begin save", "leagues", len(leagues), "teams", len(teams), "events", len(events), "lines", len(tickets))
	err = c.store.SaveLeagues(ctx, leagues)
	c.errLog(err)
	err = c.store.SaveTeams(ctx, teams)
	c.errLog(err)
	err = c.store.SaveEvents(ctx, events)
	c.errLog(err)
	err = c.store.SaveLines(ctx, tickets)
	c.errLog(err)

	//c.log.Info("begin collect lines")
	//lines, err := c.CollectLines(events)
	//c.errLog(err)
	//c.log.Infow("CollectLines done", "count", len(lines))
	//err = c.store.SaveLines(ctx, lines)
	//c.errLog(err)
	return nil
}

func (c *Collector) CollectEvents(ctx context.Context, sportId int64, dateType string, leagues *[]store.League, teams *[]store.Team, events *[]store.Event, tickets *[]store.Ticket) error {
	startRequest := time.Now()
	req := api.ShowAllOddsRequest{GameId: sportId, DateType: api.DateTypeEnum(dateType), BetTypeClass: api.OU}
	clientLock.Lock()
	resp, _, err := c.client.OddsApi.GetOdds(ctx, "BF").ShowAllOddsRequest(req).Execute()
	clientLock.Unlock()
	if c.errLogAndSleep(err, startRequest) || resp.ErrorCode != 0 {
		return errors.Errorf("request_error")
	}
	data := resp.GetData()
	for leagueId, leagueName := range data.GetLeagueN() {
		id, err := strconv.ParseInt(leagueId, 10, 64)
		if err != nil {
			c.log.Infow("parse_leagueId_error", "origin", leagueId)
			continue
		}
		*leagues = append(*leagues, store.League{Id: id, Name: help.ClearName(leagueName), SportId: sportId})
	}
	for teamId, teamName := range data.GetTeamN() {
		id, err := strconv.ParseInt(teamId, 10, 64)
		if err != nil {
			c.log.Infow("parse_teamId_error", "origin", teamId)
			continue
		}
		*teams = append(*teams, store.Team{Id: id, Name: help.ClearName(teamName), SportId: sportId})
	}
	for i := range data.GetNewMatch() {
		//c.log.Infow("event", "event", event)
		newEvent := store.Event{
			Id:         data.GetNewMatch()[i].GetMatchId(),
			Home:       data.GetNewMatch()[i].GetTeamId1(),
			Away:       data.GetNewMatch()[i].GetTeamId2(),
			LeagueId:   data.GetNewMatch()[i].GetLeagueId(),
			SportId:    data.GetNewMatch()[i].GetGameID(),
			EventState: string(data.GetNewMatch()[i].GetMaT()),
			Starts:     help.ConvertTimeZone(data.GetNewMatch()[i].GetGameTime()),
		}
		*events = append(*events, newEvent)
		if dateType == "e" {
			//c.log.Info("got early event")
			//c.log.Infow("earlyEventIn", "event", newEvent)
			earlyEventChan <- newEvent
		}
		markets := data.GetNewMatch()[i].GetMarkets()
		for i := range markets {
			*tickets = append(*tickets, store.Ticket{Id: markets[i].GetMarketId(),
				BetTypeId: markets[i].GetBetTypeId(),
				Points:    markets[i].Line,
				EventId:   markets[i].GetMatchId(),
				Cat:       markets[i].GetCat(),
			})
		}
	}
	return nil
}

func (c *Collector) CollectMoreEvents(ctx context.Context, sportId int64, tickets *[]store.Ticket) error {
	startRequest := time.Now()
	req := api.ShowAllOddsRequest{GameId: sportId, DateType: api.T, BetTypeClass: api.MORE}
	clientLock.Lock()
	resp, _, err := c.client.OddsApi.GetOdds(ctx, "BF").ShowAllOddsRequest(req).Execute()
	clientLock.Unlock()
	if c.errLogAndSleep(err, startRequest) || resp.ErrorCode != 0 {
		return errors.Errorf("request_error")
	}
	data := resp.GetData()
	//c.log.Infow("", "", data)
	for i := range data.GetNewMatch() {
		markets := data.GetNewMatch()[i].GetMarkets()

		for i := range markets {
			//c.log.Infow("", "", markets[i])
			*tickets = append(*tickets, store.Ticket{Id: markets[i].GetMarketId(),
				BetTypeId: markets[i].GetBetTypeId(),
				Points:    markets[i].Line,
				EventId:   markets[i].GetMatchId(),
				Cat:       markets[i].GetCat(),
			})
		}
	}
	return nil
}

func (c *Collector) CollectEarlyEvents() {
	var tickets []store.Ticket
	var ticketLock sync.Mutex
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			//c.log.Infow("got_ticker", "t", t)
			ticketLock.Lock()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := c.store.SaveLines(ctx, tickets)
			cancel()
			if err != nil {
				c.log.Info(err)
			}
			tickets = nil
			ticketLock.Unlock()
		case event := <-earlyEventChan:
			_, b := c.store.Cache.Get(event.Id)
			if b {
				//c.log.Infow("event_in_cache", "eventId", event.Id)
				break
			}
			starts, err := time.Parse(time.RFC3339, event.Starts)
			if err != nil {
				c.log.Infof("parse_start_time_error: %q", event.Starts)
				break
			}
			if starts.Sub(time.Now()) > earlyEventStartLimit {
				//c.log.Infow("event_too_early", "starts", starts, "eventId", event.Id)
				break
			}
			//c.log.Infow("got_earlyEvent", "event", event.Id, "starts", event.Starts)
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			lines, err := c.processEarlyEvent(ctx, event)
			cancel()
			if err != nil {
				c.log.Info(err)
				break
			}
			ticketLock.Lock()
			tickets = append(tickets, lines...)
			ticketLock.Unlock()
			c.store.Cache.SetWithTTL(event.Id, true, 1, time.Hour)
			//c.log.Infow("done_earlyEvent", "event", event.Id, "starts", event.Starts)
		}
	}
}

func (c *Collector) processEarlyEvent(ctx context.Context, event store.Event) ([]store.Ticket, error) {
	var tickets []store.Ticket
	lines, err := c.CollectLines(ctx, event.SportId, event.Id, event.EventState, "OU")
	if err != nil {
		return nil, err
	}
	tickets = append(tickets, lines...)
	lines, err = c.CollectLines(ctx, event.SportId, event.Id, event.EventState, "more")
	if err != nil {
		return nil, err
	}
	tickets = append(tickets, lines...)
	return tickets, nil
}
func (c *Collector) CollectLines(ctx context.Context, sportId int64, eventId int64, eventState string, betType string) ([]store.Ticket, error) {
	var tickets []store.Ticket
	req := api.ShowAllOddsRequest{GameId: sportId,
		DateType:     api.DateTypeEnum(eventState),
		BetTypeClass: api.BetTypeClassEnum(betType),
		Matchid:      &eventId}
	clientLock.Lock()
	resp, _, err := c.client.OddsApi.GetMarkets(ctx, "BF").ShowAllOddsRequest(req).Execute()
	clientLock.Unlock()
	if err != nil {
		return nil, err
	}
	if resp.ErrorCode != 0 {
		return nil, errors.Errorf("got_error_code %v", resp.ErrorCode)
	}
	data := resp.GetData()
	markets := data.GetNewOdds()
	if len(markets) == 0 {
		lines := data.GetMarkets()
		markets = lines.GetNewOdds()
	}
	for i := range markets {
		if len(markets[i].GetSelections()) > 3 {
			continue
		}
		tickets = append(tickets, store.Ticket{
			Id:        markets[i].GetMarketId(),
			BetTypeId: markets[i].GetBetTypeId(),
			Points:    markets[i].Line,
			EventId:   markets[i].GetMatchId(),
			Cat:       markets[i].GetCat(),
		})
	}
	return tickets, nil
}
