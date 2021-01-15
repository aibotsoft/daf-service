package handler

import (
	"context"
	"fmt"
	"github.com/aibotsoft/daf-service/pkg/store"
	api "github.com/aibotsoft/gen/dafapi"
	"github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/status"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

//var BetTypeStopList = []int64{
//	4,
//	30,
//	152,
//	158,
//	165,
//	166,
//	392,
//	399,
//	405,
//	413,
//	414,
//	416,
//	703,
//	1204,
//	1235,
//	1302,
//	1317,
//	1325,
//	2707,
//}

//func checkStopList(betTypeId int64) bool {
//	for i := range BetTypeStopList {
//		if BetTypeStopList[i] == betTypeId {
//			return true
//		}
//	}
//	return false
//}
func betTypeFromId(betTypeId int64) string {
	for i := range BetTypeOUList {
		if BetTypeOUList[i] == betTypeId {
			return "OU"
		}
	}
	return "more"
}

var BetTypeOUList = []int64{1, 2, 3, 5, 7, 8, 12, 15, 20, 21, 153, 154, 156, 9001}

func (h *Handler) collectLines(ctx context.Context, side *fortedpb.SurebetSide, line *store.Ticket) error {
	ctxAuth, err := h.auth.Auth(ctx)
	if err != nil {
		side.Check.Status = status.BadBettingStatus
		return err
	}
	eventId, _ := strconv.ParseInt(side.EventId, 10, 64)

	req := api.ShowAllOddsRequest{
		GameId:       side.SportId,
		DateType:     api.DateTypeEnum(line.EventState),
		BetTypeClass: api.BetTypeClassEnum(betTypeFromId(line.BetTypeId)),
		Matchid:      &eventId,
	}
	key := fmt.Sprintf("sport:%v:EventState:%v:betType:%v:eventId:%v", side.SportId, line.EventState, betTypeFromId(line.BetTypeId), eventId)
	_, b := h.store.Cache.Get(key)
	if b {
		h.log.Infow("lines_collected_recently", "key", key)
		return nil
	}
	resp, _, err := h.client.OddsApi.GetMarkets(ctxAuth, "").ShowAllOddsRequest(req).Execute()
	if err != nil {
		side.Check.Status = status.StatusNotFound
		return err
	}
	if resp.ErrorCode != 0 {
		return errors.Errorf("got_error_code %v", resp.ErrorCode)
	}
	var tickets []store.Ticket
	data := resp.GetData()
	markets := data.GetNewOdds()
	if len(markets) == 0 {
		lines := data.GetMarkets()
		markets = lines.GetNewOdds()
	}
	if len(markets) == 0 {
		h.log.Infow("no_markets", "resp", resp, "key", key)
	}
	for i := range markets {
		if len(markets[i].GetSelections()) > 3 {
			continue
		}
		t := store.Ticket{
			Id:        markets[i].GetMarketId(),
			BetTypeId: markets[i].GetBetTypeId(),
			Points:    markets[i].Line,
			EventId:   markets[i].GetMatchId(),
			Cat:       markets[i].GetCat(),
		}
		h.log.Infow("", "ticket", t)
		tickets = append(tickets, t)
	}
	err = h.store.SaveLines(ctx, tickets)
	if err != nil {
		side.Check.Status = status.StatusNotFound
		return err
	}
	h.store.Cache.SetWithTTL(key, true, 1, time.Second*30)
	return nil
}

func (h *Handler) processLine(ctx context.Context, side *fortedpb.SurebetSide, line *store.Ticket) error {
	if side.MarketId == 0 {
		if side.MarketName == "Ф1(0)" || side.MarketName == "Ф2(0)" {
			foundLine, err := h.store.FindLine(ctx, DrawNoBet, side.EventId)
			if err != nil {
				side.Check.Status = status.StatusError
				h.log.Info(err)
				return err
			}
			if len(foundLine) == 1 {
				h.log.Info("DrawNoBet:", foundLine)
				side.MarketId = foundLine[0].Id
				line.Id = side.MarketId
				return nil
			}
		} else if side.MarketName == "1/4 Ф1(0)" || side.MarketName == "1/4 Ф2(0)" {
			foundLine, err := h.store.FindLine(ctx, _1QuarterMoneyline, side.EventId)
			if err != nil {
				side.Check.Status = status.StatusError
				h.log.Info(err)
				return err
			}
			if len(foundLine) == 1 {
				h.log.Info("_1QuarterMoneyline:", foundLine)
				side.MarketId = foundLine[0].Id
				line.Id = side.MarketId
				return nil
			}
		}

		foundLine, err := h.store.FindLine(ctx, line.BetTypeId, side.EventId)
		if err != nil {
			side.Check.Status = status.StatusError
			h.log.Info(err)
			return err
		}
		if len(foundLine) == 0 {
			h.log.Infow("not_found_line_in_db_begin_request", "betTypeId", line.BetTypeId, "m", side.MarketName, "eventId", side.EventId,
				"home", side.Home, "away", side.Away, "sportId", side.SportId)
			err := h.collectLines(ctx, side, line)
			if err != nil {
				return err
			}
		}
		foundLine, err = h.store.FindLine(ctx, line.BetTypeId, side.EventId)
		if err != nil {
			side.Check.Status = status.StatusError
			h.log.Info(err)
			return err
		}
		switch len(foundLine) {
		case 0:
			side.Check.Status = status.StatusNotFound
			side.Check.StatusInfo = "not_found_line_in_db"
			h.log.Infof("not_found_line_in_db: BetTypeId: %v, EventId: %v, Points: %v, MarketName: %v, Sport: %v",
				line.BetTypeId, side.EventId, line.GetPoints(), side.MarketName, side.SportName)
			return errors.New(side.Check.StatusInfo)

		case 1:
			if line.ComparePoints(foundLine[0].Points) {
				side.MarketId = foundLine[0].Id
				line.Id = side.MarketId
				return nil
			} else {
				side.Check.Status = status.HandicapChanged
				side.Check.StatusInfo = fmt.Sprintf("handicap_changed_from: %v to %v", line.GetPoints(), foundLine[0].GetPoints())
				h.log.Infof("points_not_equal: BetTypeId: %v, EventId: %v, Points: %v, MarketName: %v, Sport: %v, PointsInDb: %v",
					line.BetTypeId, side.EventId, line.GetPoints(), side.MarketName, side.SportName, foundLine[0].GetPoints())
				h.log.Info("begin_delete_single_changed_line:", foundLine[0].Id)
				err := h.store.DeleteLine(ctx, foundLine[0].Id)
				if err != nil {
					h.log.Info("delete_line_error: ", err)
				}
				return errors.New(status.HandicapChanged)
			}
		default:
			for i := range foundLine {
				if line.ComparePoints(foundLine[i].Points) {
					if line.CompareCat(foundLine[i].Cat) {
						side.MarketId = foundLine[i].Id
						line.Id = side.MarketId
						return nil
					}
				}
			}

			var pointList []float64
			for i := range foundLine {
				pointList = append(pointList, foundLine[i].GetPoints())
			}
			h.log.Infow("not_found_points", "needPoints", line.GetPoints(), "gotInDb", pointList,
				"sport", side.SportName, "league", side.LeagueName, "home", side.Home, "away", side.Away, "eventId", side.EventId, "market", side.MarketName,
				"BetTypeId", line.BetTypeId, "EventState", line.EventState)

			err := h.collectLines(ctx, side, line)
			if err != nil {
				h.log.Info(err)
			}
			side.Check.Status = status.StatusNotFound
			side.Check.StatusInfo = "not_found_points_in_db"
			return errors.Errorf("not_found_line_in_db: BetTypeId: %v, EventId: %v, Points: %v, MarketName: %v", line.BetTypeId, side.EventId, line.GetPoints(), side.MarketName)
		}
	}
	line.Id = side.MarketId
	return nil
}
