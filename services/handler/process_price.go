package handler

import (
	"context"
	"fmt"
	"github.com/aibotsoft/daf-service/pkg/store"
	api "github.com/aibotsoft/gen/dafapi"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/status"
	"github.com/aibotsoft/micro/util"
	"strconv"
	"time"
)

const priceCacheTime = 2800 * time.Millisecond

var timeLock = NewTimeoutLock(460 * time.Millisecond)

func priceKey(eventId string, lineId int64, betTypeId int64, betTeam string) string {
	return fmt.Sprintf("price_key:%v:%d:%d:%v", eventId, lineId, betTypeId, betTeam)
}
func (h *Handler) GetPrice(ctx context.Context, side *pb.SurebetSide, line *store.Ticket) error {
	key := priceKey(side.EventId, line.Id, line.BetTypeId, line.BetTeam)
	got, b := h.store.Cache.Get(key)
	var ticket api.TicketItem
	if !b {
		timeLock.Take()
		defer timeLock.Release()
		ctxAuth, err := h.auth.Auth(ctx)
		if err != nil {
			side.Check.Status = status.BadBettingStatus
			side.Check.StatusInfo = "service_not_active"
			return nil
		}
		eventId, _ := strconv.ParseInt(side.EventId, 10, 64)
		//h.log.Debugw("begin_ticket_req", "type", api.OU, "matchId", eventId, "betteam", line.BetTeam,
		//	"Oddsid", line.Id, "Bettype", line.BetTypeId, "sportId", side.SportId)
		resp, _, err := h.client.TicketsApi.GetTickets(ctxAuth).
			ItemList0Type(api.OU).
			ItemList0Matchid(eventId).
			ItemList0Betteam(line.BetTeam).
			ItemList0Oddsid(line.Id).
			ItemList0Bettype(line.BetTypeId).
			ItemList0Gameid(side.SportId).Execute()
		if err != nil {
			h.log.Info("GetTickets_error", err)
			return err
		}

		if resp.GetErrorCode() != 0 {
			side.Check.Status = status.StatusNotFound
			side.Check.StatusInfo = fmt.Sprintf("service_error_code: %v", resp.GetErrorCode())
			h.log.Info(resp)
			return nil
		}

		data := resp.GetData()
		if len(data) != 1 {
			side.Check.Status = status.StatusNotFound
			h.log.Infow("ticket_not_one", "data", data)
			return nil
		}
		//h.log.Debugw("got_ticket", "data", data)
		ticket = data[0]
		code := ticket.GetCode()
		common := ticket.GetCommon()
		if code == 6 {
			side.Check.Status = status.MarketClosed
			side.Check.StatusInfo = common.GetErrorMsg()
			h.log.Infow("get_price", "resp", resp)
			return nil
		}
		if common.GetErrorCode() == 2 {
			side.Check.Status = status.Suspended
			side.Check.StatusInfo = fmt.Sprintf("suspended, ErrorCode:%v, ErrorMsg:%v", common.GetErrorCode(), common.GetErrorMsg())
			//h.log.Infow("suspended", "ticket", ticket)
			return nil
		}
		h.store.Cache.SetWithTTL(key, ticket, 1, priceCacheTime)
	} else {
		h.log.Debugw("got_ticket_from_cache", "key", key)
		ticket = got.(api.TicketItem)
	}
	minBet, err := strconv.ParseFloat(ticket.GetMinbet(), 64)
	if err != nil {
		h.log.Infow("parse_minBet_error", "minBet", ticket.GetMinbet())
		side.Check.Status = status.StatusNotFound
		side.Check.StatusInfo = "parse_minBet_error"
		return nil
	}
	maxBet, err := strconv.ParseFloat(ticket.GetMaxbet(), 64)
	if err != nil {
		h.log.Infow("parse_maxBet_error", "maxBet", ticket.GetMaxbet())
		side.Check.Status = status.StatusNotFound
		side.Check.StatusInfo = "parse_maxBet_error"
		return nil
	}
	price, err := strconv.ParseFloat(ticket.GetDisplayOdds(), 64)
	if err != nil {
		h.log.Infow("parse_maxBet_error", "DisplayOdds", ticket.GetDisplayOdds())
		side.Check.Status = status.StatusNotFound
		side.Check.StatusInfo = "parse_price_error"
		return nil
	}
	if price == 0 {
		side.Check.Status = status.StatusNotFound
		side.Check.StatusInfo = "price_is_0"
		h.log.Infow("ticket", "ticket", ticket)
		return nil
	}
	if line.Points != nil {
		if line.GetSidePoints() != ticket.GetLine() {
			side.Check.Status = status.HandicapChanged
			side.Check.StatusInfo = fmt.Sprintf("handicap_shoold_be: %v, is: %v", line.GetSidePoints(), ticket.GetLine())
			h.log.Infow("HandicapChanged", "ticket", ticket)
			h.log.Info("begin_delete_changed_line:", line.Id)
			err := h.store.DeleteLine(ctx, line.Id)
			if err != nil {
				h.log.Info("delete_line_error: ", err)
			}
			return nil
		}
	}
	tickets.Store(side.Check.Id, ticket)

	side.Check.Status = status.StatusOk
	side.Check.MinBet = util.ToUSD(minBet, side.Check.Currency)
	side.Check.MaxBet = util.ToUSD(maxBet, side.Check.Currency)
	side.Check.Price = price
	side.Check.Balance = util.ToUSDInt(h.GetBalance(), side.Check.Currency)
	side.Check.FillFactor = h.balance.CalcFillFactor()
	side.Check.SubService = "daf"
	return nil
}
