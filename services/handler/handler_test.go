package handler

import (
	"context"
	"github.com/aibotsoft/daf-service/pkg/api"
	"github.com/aibotsoft/daf-service/pkg/store"
	"github.com/aibotsoft/daf-service/services/auth"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/stretchr/testify/assert"
	"testing"
)

var h *Handler

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	db := sqlserver.MustConnectX(cfg)
	serviceApi := api.New(cfg, log)
	sto := store.NewStore(cfg, log, db)
	conf := config_client.New(cfg, log)
	au := auth.New(cfg, log, sto, serviceApi, conf)
	h = NewHandler(cfg, log, serviceApi, sto, au, conf)
	m.Run()
	h.Close()
}
func sbHelper(t *testing.T) *pb.Surebet {
	t.Helper()
	return &pb.Surebet{
		CreatedAt:       "",
		Starts:          "",
		FortedHome:      "",
		FortedAway:      "",
		FortedProfit:    0,
		FortedSport:     "",
		FortedLeague:    "",
		FilterName:      "",
		SkynetId:        0,
		FortedSurebetId: 0,
		SurebetId:       0,
		LogId:           0,
		Calc:            nil,
		Currency:        []pb.Currency{{Code: "USD", Value: 1}, {Code: "EUR", Value: 0.93}},
		Members: []*pb.SurebetSide{{
			ServiceName: "Dafabet",
			SportName:   "Table Tennis",
			LeagueName:  "Moscow Liga Pro Men Singles (Set Handicap)",
			Home:        "Andrey Menshikov",
			Away:        "Ilya Novikov",
			MarketName:  "П2",
			MarketId:    0,
			Price:       0,
			PriceId:     0,
			EventId:     0,
			Check: &pb.Check{
				Id:           0,
				AccountId:    0,
				AccountLogin: "",
				Status:       "",
				StatusInfo:   "",
				CountLine:    0,
				CountEvent:   0,
				AmountEvent:  0,
				AmountLine:   0,
				MinBet:       0,
				MaxBet:       0,
				Balance:      0,
				Price:        0,
				Currency:     0,
				Done:         0,
			},
			BetConfig: nil,
			CheckCalc: nil,
			ToBet:     nil,
			Bet:       nil,
		}},
	}
}
func TestHandler_CheckLine(t *testing.T) {
	err := h.CheckLine(context.Background(), sbHelper(t))
	assert.NoError(t, err)
}

//func TestHandler_FindTicket(t *testing.T) {
//	err := h.FindTicket(context.Background(), sbHelper(t))
//	assert.NoError(t, err)
//}

func TestHandler_PlaceBet(t *testing.T) {
	err := h.PlaceBet(context.Background(), sbHelper(t))
	assert.NoError(t, err)
}
