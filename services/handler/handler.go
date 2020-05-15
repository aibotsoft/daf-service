package handler

import (
	"context"
	"github.com/aibotsoft/daf-service/pkg/api"
	"github.com/aibotsoft/daf-service/pkg/store"
	"github.com/aibotsoft/daf-service/services/auth"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/status"
	"github.com/aibotsoft/micro/util"
	"go.uber.org/zap"
)

type Handler struct {
	cfg     *config.Config
	log     *zap.SugaredLogger
	api     *api.Api
	store   *store.Store
	auth    *auth.Auth
	Conf    *config_client.ConfClient
	balance Balance
}

func NewHandler(cfg *config.Config, log *zap.SugaredLogger, api *api.Api, store *store.Store, auth *auth.Auth, conf *config_client.ConfClient) *Handler {
	h := &Handler{cfg: cfg, log: log, api: api, store: store, auth: auth, Conf: conf, balance: Balance{}}
	h.GetBalance()
	return h
}

func (h *Handler) Close() {
	h.store.Close()
	h.Conf.Close()
}

var eventStates = []string{"today", "earlyMarket"}

func (h *Handler) CheckLine(ctx context.Context, sb *pb.Surebet) error {
	side := sb.Members[0]
	side.Check.AccountId = h.auth.Account.Id
	side.Check.AccountLogin = h.auth.Account.Username
	side.Check.Currency = h.GetCurrency(sb)
	line, err := Convert(side.MarketName)
	if err != nil {
		h.log.Infow(err.Error(), "sport", side.SportName, "league", side.LeagueName, "home", side.Home, "away", side.Away)
		side.Check.StatusInfo = "not converted market: " + side.MarketName
		return nil
	}
	err = h.store.GetStat(side)
	if err != nil {
		h.log.Error(err)
		side.Check.StatusInfo = "error get stat for event"
		return nil
	}
	err = h.processSport(ctx, side)
	if err != nil {
		side.Check.StatusInfo = "not found sport"
		h.log.Info(err)
		return nil
	}
	err = h.processLeague(ctx, side)
	if err != nil {
		side.Check.StatusInfo = "not found league"
		h.log.Info(err)
		return nil
	}
	err = h.processEvent(ctx, side)
	if err != nil {
		side.Check.StatusInfo = "not found event"
		h.log.Info(err)
		return nil
	}
	err = h.processLine(ctx, side, line)
	if err != nil {
		side.Check.StatusInfo = "not found line"
		h.log.Info(err)
		return nil
	}
	err = h.api.GetTicket(ctx, h.auth.Base(), line)
	switch err {
	case nil:
		side.Check.Status = status.StatusOk
		side.Check.MinBet = util.ToUSD(line.MinBet, side.Check.Currency)
		side.Check.MaxBet = util.ToUSD(line.MaxBet, side.Check.Currency)
		side.Check.Price = line.Price
		side.Check.Balance = util.ToUSDInt(h.GetBalance(), side.Check.Currency)
	case api.DiffPointsError:
		side.Check.Status = status.StatusNotFound
		side.Check.StatusInfo = "points changed"
	default:
		side.Check.Status = status.StatusError
	}
	h.log.Infow("check done", "sportId", side.SportId, "leagueId", side.LeagueId, "eventId", side.EventId, "marketId", side.MarketId,
		//"MarketName", side.MarketName, "SportName", side.SportName,	"LeagueName", side.LeagueName, "Home", side.Home, "Away", side.Away,
		"line", line, "check", side.Check)
	//h.log.Infow("check done","check", side.Check)
	return nil
}

func (h *Handler) PlaceBet(ctx context.Context, sb *pb.Surebet) error {
	//side := sb.Members[0]
	//go func() {
	//	err := h.store.SaveCheck(sb)
	//	if err != nil {
	//		h.log.Error(err)
	//	}
	//}()
	//err := h.store.SaveCheck(sb)
	//stake := util.AdaptStake(side.CheckCalc.Stake, side.Check.Currency, side.BetConfig.RoundValue)
	stake := float64(1)
	h.log.Infow("begin place bet", "stake", stake)
	ticket := &api.Ticket{
		Id:        250615816,
		BetTeam:   "a",
		BetTypeId: 20,
	}

	placeResult, err := h.api.PlaceBet(ctx, h.auth.Base(), ticket, stake)
	if err != nil {
		h.log.Error(err)
		return nil
	}

	//h.log.Info(side)
	//h.log.Info(stake)
	h.log.Info(placeResult)
	return err
}

func (h *Handler) GetResults(ctx context.Context) ([]pb.BetResult, error) {
	return []pb.BetResult{}, nil
}
