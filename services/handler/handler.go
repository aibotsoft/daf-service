package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/aibotsoft/daf-service/pkg/store"
	"github.com/aibotsoft/daf-service/services/auth"
	api "github.com/aibotsoft/gen/dafapi"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/status"
	"github.com/aibotsoft/micro/util"
	"go.uber.org/zap"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var BadBettingStatusError = errors.New("BadBettingStatus")
var tickets sync.Map
var CheckLineError = errors.New("CheckLineError")

const LockTimeOut = time.Second * 28

type Handler struct {
	cfg     *config.Config
	log     *zap.SugaredLogger
	client  *api.APIClient
	store   *store.Store
	auth    *auth.Auth
	Conf    *config_client.ConfClient
	balance Balance
}

func NewHandler(cfg *config.Config, log *zap.SugaredLogger, store *store.Store, auth *auth.Auth, conf *config_client.ConfClient) *Handler {
	clientConfig := api.NewConfiguration()
	clientConfig.Debug = cfg.Service.Debug
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse("https://ismart.dafabet.com")
	jar.SetCookies(u, []*http.Cookie{{Name: "SkinMode", Value: "3"}})

	tr := &http.Transport{
		TLSHandshakeTimeout: 0 * time.Second,
		IdleConnTimeout:     0 * time.Second,
		MaxIdleConnsPerHost: 20,
		MaxConnsPerHost:     4,
		MaxIdleConns:        20,
	}
	clientConfig.HTTPClient = &http.Client{Transport: tr, Jar: jar}
	client := api.NewAPIClient(clientConfig)

	h := &Handler{cfg: cfg, log: log, client: client, store: store, auth: auth, Conf: conf}
	return h
}

func (h *Handler) Close() {
	h.store.Close()
	h.Conf.Close()
}

func (h *Handler) RaceDebug(id int64) bool {
	res := h.store.SetVerifyWithTTL("key", id, time.Minute)
	return res
}

func (h *Handler) GetLock(sb *pb.Surebet) bool {
	side := sb.Members[0]
	key := side.Market.BetType + side.Url
	for i := 0; i < 40; i++ {
		got, b := h.store.Cache.Get(key)
		if b && got.(int64) != sb.SurebetId {
			time.Sleep(time.Millisecond * 50)
		} else {
			return h.store.SetVerifyWithTTL(key, sb.SurebetId, LockTimeOut)
		}
	}
	return false
}

func (h *Handler) ReleaseCheck(ctx context.Context, sb *pb.Surebet) {
	side := sb.Members[0]
	key := side.Market.BetType + side.Url
	got, b := h.store.Cache.Get(key)
	if b && got.(int64) == sb.SurebetId {
		h.store.Cache.Del(key)
	}
}
func (h *Handler) CheckLine(ctx context.Context, sb *pb.Surebet) error {
	side := sb.Members[0]
	side.Check.AccountId = h.auth.Account.Id
	side.Check.AccountLogin = h.auth.Account.Username
	side.Check.Currency = h.GetCurrency(sb)
	lockStart := time.Now()
	ok := h.GetLock(sb)
	if !ok {
		side.Check.Status = status.ServiceBusy
		side.Check.StatusInfo = "service_already_check_this_market"
		h.log.Debugw(status.ServiceBusy, "my_surebetId", sb.SurebetId)
		return nil
	}
	lockTime := time.Since(lockStart)

	line, err := Convert(side)
	if err != nil {
		h.log.Infow(err.Error(), "sport", side.SportName, "league", side.LeagueName, "home", side.Home, "away", side.Away)
		side.Check.Status = status.StatusNotFound
		side.Check.StatusInfo = "not_converted_market: " + side.MarketName
		return nil
	}
	err = h.ParseUrl(side, line)
	if err != nil {
		h.log.Error(err)
		return nil
	}

	err = h.processLine(ctx, side, line)
	if err != nil {
		return nil
	}
	err = h.GetPrice(ctx, side, line)
	if err != nil {
		h.log.Info(err)
		side.Check.Status = status.StatusError
		return nil
	}
	err = h.store.GetStat(side)
	if err != nil {
		h.log.Error(err)
		side.Check.Status = status.StatusError
		side.Check.StatusInfo = "get_stat_for_event_error"
		return nil
	}
	err = h.store.GetStartTime(ctx, side)
	if err != nil {
		h.log.Info("get_start_time_error: ", err, " eventId:", side.EventId)
		side.Starts = sb.Starts
	}

	//h.log.Debug(side.Starts)
	//h.log.Infow("", "line", line, "sportId", side.SportId, "leagueId", side.LeagueId, "eventId", side.EventId, "homeId", side.HomeId, "awayId", side.AwayId)
	//h.log.Infow("check done", "sportId", side.SportId, "leagueId", side.LeagueId, "eventId", side.EventId, "marketId", side.MarketId,
	//	//"MarketName", side.MarketName, "SportName", side.SportName,	"LeagueName", side.LeagueName, "Home", side.Home, "Away", side.Away,
	//	"line", line, "check", side.Check)
	if side.Check.Status != status.StatusOk {
		h.log.Infow("check_not_ok",
			"check", side.Check,
			"lineId", line.Id,
			"BetTeam", line.BetTeam,
			"BetTypeId", line.BetTypeId,
			"Points", line.Points,
			"line_cat", line.Cat,
			"sportId", side.SportId,
			"leagueId", side.LeagueId,
			"eventId", side.EventId,
			"homeId", side.HomeId,
			"awayId", side.AwayId,
			"m", side.MarketName)
	} else {
		h.log.Infow("check_ok",
			"p", side.Check.Price,
			"fp", side.Price,
			"max", side.Check.MaxBet,
			//"lineId", line.Id,
			//"BetTeam", line.BetTeam,
			//"BetTypeId", line.BetTypeId,
			"points", line.Points,
			"sportId", side.SportId,
			"leagueId", side.LeagueId,
			"eventId", side.EventId,
			"m", side.MarketName,
			"lock_time", lockTime,
			//"homeId", side.HomeId,
			//"awayId", side.AwayId,
		)
	}
	return nil
}

func (h *Handler) PlaceBet(ctx context.Context, sb *pb.Surebet) error {
	side := sb.Members[0]
	side.Bet.Status = status.StatusNotAccepted
	//go func() {
	//	err := h.store.SaveCheck(sb)
	//	if err != nil {
	//		h.log.Error(err)
	//	}
	//}()
	err := h.store.SaveCheck(sb)
	if err != nil {
		h.log.Info("save_check_error: ", err)
		side.Bet.Status = status.StatusError
		side.Bet.StatusInfo = "save_check_error"
		return nil
	}

	stake := util.AdaptStake(side.CheckCalc.Stake, side.Check.Currency, side.BetConfig.RoundValue)
	ctxAuth, err := h.auth.Auth(ctx)
	if err != nil {
		side.Bet.Status = status.BadBettingStatus
		side.Bet.StatusInfo = "service_not_active"
		return nil
	}
	got, ok := tickets.Load(side.Check.Id)
	if !ok {
		side.Bet.Status = status.StatusError
		side.Bet.StatusInfo = "check_ticket_not_found"
		return nil
	}
	ticket := got.(api.TicketItem)
	h.log.Debugw("got_ticket_to_bet", "ticket", ticket)
	betType, err := strconv.ParseInt(ticket.GetBettype(), 10, 64)
	if err != nil {
		h.log.Info("parse Bettype error: ", err)
		side.Bet.Status = status.StatusError
		return nil
	}
	odds, err := strconv.ParseFloat(ticket.GetDisplayOdds(), 32)
	if err != nil {
		h.log.Info("parse odds error: ", err)
	}
	resp, _, err := h.client.TicketsApi.PlaceBet(ctxAuth).
		ItemList0Type(ticket.GetTicketType()).
		ItemList0Bettype(betType).
		ItemList0Oddsid(ticket.GetOddsID()).
		ItemList0Odds(float32(odds)).
		//ItemList0Line(ticket.GetLine()).
		ItemList0Matchid(ticket.GetMatchid()).
		ItemList0Betteam(ticket.GetBetteam()).
		ItemList0Stake(int64(stake)).
		ItemList0QuickBet(ticket.GetQuickBet()).
		ItemList0ChoiceValue(ticket.GetChoiceValue()).
		ItemList0Home(ticket.GetHomeName()).
		ItemList0Away(ticket.GetAwayName()).
		ItemList0Gameid(ticket.GetSportType()).
		ItemList0IsInPlay(ticket.GetIsLive()).
		ItemList0Hdp1(ticket.GetHdp1()).
		ItemList0Hdp2(ticket.GetHdp2()).
		Execute()
	if err != nil {
		h.log.Info("place_bet_error: ", err)
		side.Bet.Status = status.StatusError
		side.Bet.StatusInfo = fmt.Sprintf("place_bet_request_error: %q", err.Error())
		return nil
	}
	if resp.ErrorCode != 0 {
		side.Bet.Status = status.StatusError
		side.Bet.StatusInfo = "service_error"
		h.log.Infow("resp", "resp", resp)
	}
	h.log.Infow("resp", "resp", resp)
	data := resp.GetData()
	itemList := data.GetItemList()
	if len(itemList) < 1 {
		h.log.Infow("bet_item_list_empty", "resp", resp)
		side.Bet.Status = status.StatusError
		side.Bet.StatusInfo = "bet_item_list_empty"
		return nil
	}
	bet := itemList[0]

	price, err := strconv.ParseFloat(bet.GetDisplayOdds(), 64)
	if err != nil {
		h.log.Info("parse_odd_error: ", err)
		price = odds
	}
	realStake, err := strconv.ParseFloat(bet.GetStake(), 64)
	if err != nil {
		h.log.Info("parse_stake_error: ", err)
		realStake = stake
	}
	if bet.GetCode() == 1 {
		side.Bet.Status = status.StatusOk
		side.Bet.StatusInfo = fmt.Sprintf("%v, TotalPerBet:%v", bet.GetMessage(), bet.GetTotalPerBet())

		side.Bet.Price = price
		side.Bet.Stake = util.ToUSD(realStake, side.Check.Currency)
		side.Bet.ApiBetId = bet.GetTransIdCash()
		h.balance.Sub(realStake)

	} else if bet.GetCode() == 0 {
		side.Bet.Status = status.PendingAcceptance
		side.Bet.StatusInfo = bet.GetMessage()
	} else if bet.GetCode() == 6 {
		side.Bet.Status = status.MarketClosed
		side.Bet.StatusInfo = bet.GetMessage()
	} else if bet.GetCode() == 7 {
		side.Bet.Status = status.AboveEventMax
		side.Bet.StatusInfo = bet.GetMessage()
	} else if bet.GetCode() == 48 {
		side.Bet.Status = status.AboveEventMax
		side.Bet.StatusInfo = bet.GetMessage()
	} else {
		side.Bet.Status = status.StatusNotAccepted
		side.Bet.StatusInfo = bet.GetMessage()
	}
	side.Bet.Done = util.UnixMsNow()
	err = h.store.SaveBet(sb)
	if err != nil {
		h.log.Error(err)
	}
	return nil
}
