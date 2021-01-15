package handler

import (
	"context"
	"database/sql"
	"github.com/aibotsoft/daf-service/pkg/help"
	"github.com/aibotsoft/daf-service/pkg/store"
	api "github.com/aibotsoft/gen/dafapi"
	"github.com/aibotsoft/gen/fortedpb"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

const eventMinPeriodSeconds = 30

var eventLast = make(map[int64]int64)
var dtList = []api.DateTypeEnum{api.E, api.T}

func (h *Handler) collectEventsForSport(ctx context.Context, sportId int64) error {
	last, ok := eventLast[sportId]
	if ok {
		if time.Now().Unix()-last < eventMinPeriodSeconds {
			return nil
		}
	}
	var leagues []store.League
	var teams []store.Team
	var eventList []store.Event
	var tickets []store.Ticket
	ctxAuth, err := h.auth.Auth(ctx)
	if err != nil {
		return err
	}
	for _, dt := range dtList {
		req := api.ShowAllOddsRequest{GameId: sportId, DateType: dt, BetTypeClass: api.OU}
		resp, _, err := h.client.OddsApi.GetOdds(ctxAuth, "").ShowAllOddsRequest(req).Execute()
		if err != nil {
			h.log.Info(err)
			return err
		}
		if resp.ErrorCode != 0 {
			h.log.Infow("got_odds_error_code", "code", resp.ErrorCode)
			continue
		}
		//c.log.Infow("", "", resp)
		data := resp.GetData()
		for leagueId, leagueName := range data.GetLeagueN() {
			id, err := strconv.ParseInt(leagueId, 10, 64)
			if err != nil {
				h.log.Infow("parse_league_id_error", "origin", leagueId)
				continue
			}
			leagues = append(leagues, store.League{Id: id, Name: help.ClearName(leagueName), SportId: sportId})
		}
		for teamId, teamName := range data.GetTeamN() {
			id, err := strconv.ParseInt(teamId, 10, 64)
			if err != nil {
				h.log.Infow("parse_team_id_error", "origin", teamId)
				continue
			}
			teams = append(teams, store.Team{Id: id, Name: help.ClearName(teamName), SportId: sportId})
		}
		events := data.GetNewMatch()
		for i := range events {
			eventList = append(eventList, store.Event{
				Id:         events[i].GetMatchId(),
				Home:       events[i].GetTeamId1(),
				Away:       events[i].GetTeamId2(),
				LeagueId:   events[i].GetLeagueId(),
				SportId:    events[i].GetGameID(),
				EventState: string(events[i].GetMaT()),
				Starts:     help.ConvertTimeZone(events[i].GetGameTime()),
			})
			markets := events[i].GetMarkets()
			for i := range markets {
				tickets = append(tickets, store.Ticket{Id: markets[i].GetMarketId(), BetTypeId: markets[i].GetBetTypeId(), Points: markets[i].Line, EventId: markets[i].GetMatchId()})
			}
		}
	}
	//h.log.Infow("resp", "", resp)
	//h.log.Infow("begin save leagues", "count", len(leagues))
	err = h.store.SaveLeagues(ctx, leagues)
	help.ErrLog(err, h.log)
	//h.log.Infow("begin save teams", "count", len(teams))
	err = h.store.SaveTeams(ctx, teams)
	help.ErrLog(err, h.log)
	//h.log.Infow("begin save events", "count", len(eventList))
	err = h.store.SaveEvents(ctx, eventList)
	help.ErrLog(err, h.log)
	//h.log.Infow("begin save lines", "count", len(tickets))
	err = h.store.SaveLines(ctx, tickets)
	help.ErrLog(err, h.log)
	eventLast[sportId] = time.Now().Unix()
	return nil
}

func (h *Handler) processEvent(ctx context.Context, side *fortedpb.SurebetSide) error {
	if side.EventId == "" {
		var err error
		side.HomeId, err = h.store.FindTeamByName(ctx, side.Home, side.SportId)
		if err != nil {
			return errors.Errorf("not_found home team: %q, sport: %q, sportId: %v", side.Home, side.SportName, side.SportId)
		}
		side.AwayId, err = h.store.FindTeamByName(ctx, side.Away, side.SportId)
		if err != nil {
			return errors.Errorf("not_found away team: %q, sport: %q, sportId: %v", side.Away, side.SportName, side.SportId)
		}
		side.EventId, err = h.store.FindEventByName(ctx, side.HomeId, side.AwayId, side.LeagueId, side.SportId)
		switch err {
		case nil:
			return nil
		case sql.ErrNoRows:
			err := h.collectEventsForSport(ctx, side.SportId)
			if err != nil {
				return err
			}
			side.EventId, err = h.store.FindEventByName(ctx, side.HomeId, side.AwayId, side.LeagueId, side.SportId)
			if err != nil {
				return errors.Errorf("not_found_event_in_db, homeId: %v, awayId: %v, leagueId: %v, sportId: %v", side.HomeId, side.AwayId, side.LeagueId, side.SportId)
			}
		default:
			return err
		}
	}
	return nil
}
