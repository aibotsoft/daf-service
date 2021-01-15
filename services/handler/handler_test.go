package handler

import (
	"context"
	"github.com/aibotsoft/daf-service/pkg/store"
	"github.com/aibotsoft/daf-service/services/auth"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/aibotsoft/micro/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var h *Handler

func TestMain(m *testing.M) {
	cfg := config.New()
	log := logger.New()
	db := sqlserver.MustConnectX(cfg)
	sto := store.NewStore(cfg, log, db)
	conf := config_client.New(cfg, log)
	a := auth.New(cfg, log, sto, conf)
	err := a.CheckAndLogin(context.Background())
	if err != nil {
		log.Error(err)
	}
	h = NewHandler(cfg, log, sto, a, conf)
	m.Run()
	h.Close()
}
func sbHelper(t *testing.T) *pb.Surebet {
	t.Helper()
	return &pb.Surebet{
		Currency: []pb.Currency{{Code: "USD", Value: 1}, {Code: "EUR", Value: 0.93}},
		Members: []*pb.SurebetSide{{
			ServiceName: "Dafabet",
			SportName:   "E-Sports",
			LeagueName:  "CLUB FRIENDLY",
			Home:        "Nam Dinh",
			Away:        "Pho Hien FC",
			MarketName:  "1/4 Ф1(0)",
			Url:         "http://12wap.mobi/underover/oddsnew.aspx?matchid=37475500&eventState=today&sport=2&leagueID=49426",
			Check:       &pb.Check{Id: util.UnixMsNow()},
			Bet:         &pb.Bet{},
			ToBet:       &pb.ToBet{Id: util.UnixMsNow(), TryCount: 0},
			BetConfig:   &pb.BetConfig{RoundValue: 1},
			CheckCalc: &pb.CheckCalc{
				Status:   "Ok",
				MaxStake: 1,
				MinStake: 1,
				MaxWin:   2,
				Stake:    1,
				Win:      1,
				IsFirst:  true,
			},
		}},
	}
}
func TestHandler_CheckLine(t *testing.T) {
	sb := sbHelper(t)
	err := h.CheckLine(context.Background(), sb)
	if assert.NoError(t, err) {
		t.Log(sb.Members[0].Check)
	}
}

func TestHandler_CheckLine_System(t *testing.T) {
	events, err := h.store.GetDemoEvents(context.Background(), 1, 43, "%beastcoast%")
	assert.NoError(t, err)
	//t.Log(events)
	for _, e := range events {
		sb := sbHelper(t)
		side := sb.Members[0]
		side.EventId = e.EventId
		side.SportId = e.SportId
		side.SportName = e.SportName
		side.LeagueName = e.LeagueName
		side.LeagueId = e.LeagueId
		side.Home = e.Home
		side.Away = e.Away
		t.Log(e)
		for _, marketName := range []string{
			//"П1",
			//"П2",
			//"1/2 П1",
			//"1/2 П2",
			"2/2 П1",
			//"2/2 П2",
			//"3/2 П1",
			//"3/2 П2",
			//"4/2 П1",
			//"4/2 П2",
			//"5/2 П1",
			//"5/2 П2",
			//"УГЛ 1",
			//"УГЛ 2",
			//"УГЛ Х",
			//"УГЛ 1/2 1",
			//"УГЛ 1/2 2",
			//"УГЛ 1/2 Х",
			//"УГЛ 2/2 1",
			//"УГЛ 2/2 2",
			//"УГЛ 2/2 Х",
			//"УГЛ 1Х",
			//"УГЛ 12",
			//"УГЛ Х2",
			//"УГЛ 1/2 1Х",
			//"УГЛ 1/2 12",
			//"УГЛ 1/2 Х2",
			//"УГЛ 2/2 1Х",
			//"УГЛ 2/2 12",
			//"УГЛ 2/2 Х2",
			//"УГЛ Чёт",
			//"УГЛ Нечёт",
			//"УГЛ 1/2 Чёт",
			//"УГЛ 1/2 Нечёт",
			//"УГЛ 2/2 Чёт",
			//"УГЛ 2/2 Нечёт",
			//"УГЛ Ф1(3,5)",
			//"УГЛ Ф2(-3,5)",
			//"УГЛ 1/2 Ф1(1,5)",
			//"УГЛ 1/2 Ф2(-1,5)",
			//"УГЛ 2/2 Ф1(0)",
			//"УГЛ 2/2 Ф2(0)",
			//"УГЛ ТМ(10)",
			//"УГЛ ТБ(10)",
			//"УГЛ 1/2 ТМ(4,5)",
			//"УГЛ 1/2 ТБ(4,5)",
			//"УГЛ 2/2 ТМ(4,5)",
			//"УГЛ 2/2 ТБ(4,5)",
			//"УГЛ ИТ1Б(4)",
			//"УГЛ ИТ1М(4)",
			//"УГЛ ИТ2Б(5,5)",
			//"УГЛ ИТ2М(5,5)",
			//"УГЛ 1/2 ИТ1Б(2)",
			//"УГЛ 1/2 ИТ1М(2)",
			//"УГЛ 1/2 ИТ2Б(2,5)",
			//"УГЛ 1/2 ИТ2М(2,5)",
			//"ЖК 1",
			//"ЖК 2",
			//"ЖК Х",
			//"ЖК 1/2 1",
			//"ЖК 1/2 2",
			//"ЖК 1/2 Х",
			//"ЖК 2/2 1",
			//"ЖК 2/2 2",
			//"ЖК 2/2 Х",
			//"ЖК 1Х",
			//"ЖК 12",
			//"ЖК Х2",
			//"ЖК 1/2 1Х",
			//"ЖК 1/2 12",
			//"ЖК 1/2 Х2",
			//"ЖК 2/2 1Х",
			//"ЖК 2/2 12",
			//"ЖК 2/2 Х2",
			//"ЖК Чёт",
			//"ЖК Нечёт",
			//"ЖК 1/2 Чёт",
			//"ЖК 1/2 Нечёт",
			//"ЖК 2/2 Чёт",
			//"ЖК 2/2 Нечёт",
			//"ЖК Ф1(-0,5)",
			//"ЖК Ф2(0,5)",
			//"ЖК 1/2 Ф1(0)",
			//"ЖК 1/2 Ф2(0)",
			//"ЖК 2/2 Ф1(0)",
			//"ЖК 2/2 Ф2(0)",
			//"ЖК ТМ(3,5)",
			//"ЖК ТБ(3,5)",
			//"ЖК 1/2 ТМ(1,5)",
			//"ЖК 1/2 ТБ(1,5)",
			//"ЖК 2/2 ТМ(1,75)",
			//"ЖК 2/2 ТБ(1,75)",

			//"1",
			//"2",
			//"Х",
			//"1/2 1",
			//"1/2 2",
			//"1/2 Х",
			//"2/2 1",
			//"2/2 2",
			//"2/2 Х",
			//"1Х",
			//"12",
			//"Х2",
			//"1/2 1Х",
			//"1/2 12",
			//"1/2 Х2",
			//"2/2 1Х",
			//"2/2 12",
			//"2/2 Х2",
			//"Чёт",
			//"Нечёт",
			//"1/2 Чёт",
			//"1/2 Нечёт",
			//"1/4 Чёт",
			//"1/4 Нечёт",
			//"2/4 Чёт",
			//"2/4 Нечёт",
			//"2/2 Чёт",
			//"2/2 Нечёт",

			//"Ф1(-2,5)",
			//"Ф2(2,5)",
			//"1/2 Ф1(-1,5)",
			//"1/2 Ф2(1,5)",
			//"2/2 Ф1(1)",
			//"2/2 Ф2(-1)",
			//"ТМ(2,5)",
			//"ТБ(2,5)",
			//"(карты) ТМ(3,5)",
			//"(карты) ТБ(3,5)",
			//"1/2 ТМ(111,5)",
			//"1/2 ТБ(111,5)",
			//"2/2 ТМ(1,75)",
			//"2/2 ТБ(1,75)",
			//"ИТ1Б(115)",
			//"ИТ1М(115)",
			//"ИТ2Б(113)",
			//"ИТ2М(113)",

			//"1/2 ИТ1Б(56,5)",
			//"1/2 ИТ1М(56,5)",
			//"1/2 ИТ2Б(55)",
			//"1/2 ИТ2М(55)",

			//"1/4 ИТ1Б(29)",
			//"1/4 ИТ1М(29)",
			//"1/4 ИТ2Б(27,5)",
			//"1/4 ИТ2М(27,5)",

			//"Старт К1",
			//"Старт К2",

		} {
			side.MarketName = marketName
			side.MarketId = 0
			side.Check.StatusInfo = ""

			err = h.CheckLine(context.Background(), sb)
			assert.NoError(t, err)
			time.Sleep(time.Millisecond * 300)
			t.Log()
		}
	}
}

//Mexico (V) -0.50@1.89
//Venezuela (V) 0.50@1.89
//[Handicap]
//Mexico (V) -0.25@1.64
//Venezuela (V) 0.25@2.16
//func TestHandler_PlaceBet(t *testing.T) {
//	ctx := context.Background()
//	events, err := h.store.GetDemoEvents(ctx, 1, 1, "%Bohemians%")
//	assert.NoError(t, err)
//	for _, e := range events[:1] {
//		sb := sbHelper(t)
//		side := sb.Members[0]
//		side.SportName = e.SportName
//		side.LeagueName = e.LeagueName
//		side.Home = e.Home
//		side.Away = e.Away
//		side.MarketName = "ТБ(2)"
//		err = h.CheckLine(ctx, sb)
//		if assert.NoError(t, err) {
//			//t.Log(side.Check)
//			err = h.PlaceBet(ctx, sb)
//			if assert.NoError(t, err) {
//				t.Log(side.Bet)
//			}
//		}
//	}
//}

func TestHandler_RaceDebug(t *testing.T) {
	go func() {
		got := h.RaceDebug(1)
		assert.True(t, got)
	}()
	go func() {
		got := h.RaceDebug(2)
		assert.True(t, got)
	}()
	time.Sleep(time.Millisecond * 20)

}
