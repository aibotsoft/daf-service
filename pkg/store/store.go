package store

import (
	"context"
	"database/sql"
	"fmt"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/cache"
	"github.com/aibotsoft/micro/config"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/dgraph-io/ristretto"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type Store struct {
	cfg   *config.Config
	log   *zap.SugaredLogger
	db    *sqlx.DB
	Cache *ristretto.Cache
}

func NewStore(cfg *config.Config, log *zap.SugaredLogger, db *sqlx.DB) *Store {
	return &Store{log: log, db: db, Cache: cache.NewCache(cfg)}
}

func (s *Store) Close() {
	err := s.db.Close()
	if err != nil {
		s.log.Error(err)
	}
	s.Cache.Close()
}
func (s *Store) SetVerifyWithTTL(key string, value interface{}, ttl time.Duration) bool {
	s.Cache.SetWithTTL(key, value, 1, ttl)
	for i := 0; i < 100; i++ {
		got, b := s.Cache.Get(key)
		if b {
			if got == value {
				return true
			} else {
				s.log.Info("got != value:", got, value)
				return false
			}
		}
		time.Sleep(time.Microsecond * 5)
	}
	return false
}

func (s *Store) SaveSession(session string, token string) error {
	_, err := s.db.Exec("dbo.uspSaveSession", sql.Named("Session", session), sql.Named("Token", token))
	return err
}
func (s *Store) LoadToken(ctx context.Context) (session string, token string, err error) {
	err = s.db.QueryRowContext(ctx, "select top 1 Session, Token from dbo.Auth order by LastCheckAt desc").Scan(&session, &token)
	if err != nil {
		return "", "", errors.Wrap(err, "load_token_error")
	}
	return
}

type Market struct {
	Id   int64
	Name string
}

func (s *Store) SaveMarkets(ctx context.Context, markets []Market) error {
	if len(markets) == 0 {
		return nil
	}
	tvp := mssql.TVP{TypeName: "MarketType", Value: markets}
	_, err := s.db.ExecContext(ctx, "dbo.uspCreateMarkets", tvp)
	return err
}

func (s *Store) FindSportByName(ctx context.Context, sportName string) (int64, error) {
	get, b := s.Cache.Get("sport:" + sportName)
	if b {
		return get.(int64), nil
	}
	var id int64
	err := s.db.GetContext(ctx, &id, "select Id from dbo.Sport where Name = @p1", sportName)
	if err != nil {
		return 0, err
	}
	s.Cache.Set("sport:"+sportName, id, 1)
	return id, nil
}
func (s *Store) FindLeagueByName(ctx context.Context, sportId int64, leagueName string) (int64, error) {
	get, b := s.Cache.Get("league:" + leagueName)
	if b {
		return get.(int64), nil
	}
	var id int64
	err := s.db.GetContext(ctx, &id, "select Id from dbo.League where Name = @p1 and SportId=@p2", leagueName, sportId)
	if err != nil {
		return 0, err
	}
	s.Cache.Set("league:"+leagueName, id, 1)
	return id, nil
}

const findEventQ = `
select Id from dbo.Event where Home = @p1 and Away=@p2 and LeagueId=@p3 and SportId=@p4 and Starts > sysdatetimeoffset()
`

func (s *Store) FindEventByName(ctx context.Context, home int64, away int64, leagueId int64, sportId int64) (id string, err error) {
	err = s.db.GetContext(ctx, &id, findEventQ, home, away, leagueId, sportId)
	return
}

//func (s *Store) FindLine(ctx context.Context, betTypeId int64, eventId int64, points *float64, cat int64) (id int64, err error) {
//	err = s.db.GetContext(ctx, &id, "dbo.uspFindLine",
//		sql.Named("EventId", eventId),
//		sql.Named("BetTypeId", betTypeId),
//		sql.Named("Points", points),
//		sql.Named("Cat", cat),
//	)
//	return
//}
const findLineQ = "select Id, BetTypeId, Points,EventId, Cat from dbo.Line where EventId = @EventId and BetTypeId = @BetTypeId"

func (s *Store) FindLine(ctx context.Context, betTypeId int64, eventId string) (lines []Ticket, err error) {
	err = s.db.SelectContext(ctx, &lines, findLineQ, sql.Named("EventId", eventId), sql.Named("BetTypeId", betTypeId))
	return
}

type Stat struct {
	MarketName  string
	CountEvent  int64
	CountLine   int64
	AmountEvent int64
	AmountLine  int64
}

func (s *Store) GetStat(side *pb.SurebetSide) error {
	var stat []Stat
	err := s.db.Select(&stat, "dbo.uspCalcStat", sql.Named("EventId", side.EventId))
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "uspCalcStat error")
	} else {
		for i := range stat {
			side.Check.AmountEvent = stat[i].AmountEvent
			side.Check.CountEvent = stat[i].CountEvent
			if stat[i].MarketName == side.MarketName {
				side.Check.CountLine = stat[i].CountLine
				side.Check.AmountLine = stat[i].AmountLine
				return nil
			}
		}
	}
	return nil
}

func (s *Store) SaveCheck(sb *pb.Surebet) error {
	side := sb.Members[0]
	_, err := s.db.Exec("dbo.uspSaveSide",
		sql.Named("Id", sb.SurebetId),
		sql.Named("SideIndex", side.Num-1),

		sql.Named("SportName", side.SportName),
		sql.Named("SportId", side.SportId),
		sql.Named("LeagueName", side.LeagueName),
		sql.Named("LeagueId", side.LeagueId),
		sql.Named("Home", side.Home),
		sql.Named("HomeId", side.HomeId),
		sql.Named("Away", side.Away),
		sql.Named("AwayId", side.AwayId),
		sql.Named("MarketName", side.MarketName),
		sql.Named("MarketId", side.MarketId),
		sql.Named("Price", side.Price),
		sql.Named("Initiator", side.Initiator),
		sql.Named("Starts", sb.Starts),
		sql.Named("EventId", side.EventId),

		sql.Named("CheckId", side.GetCheck().GetId()),
		sql.Named("AccountLogin", side.GetCheck().GetAccountLogin()),
		sql.Named("CheckPrice", side.GetCheck().GetPrice()),
		sql.Named("CheckStatus", side.GetCheck().GetStatus()),
		sql.Named("CountLine", side.GetCheck().GetCountLine()),
		sql.Named("CountEvent", side.GetCheck().GetCountEvent()),
		sql.Named("AmountEvent", side.GetCheck().GetAmountEvent()),
		sql.Named("AmountLine", side.GetCheck().GetAmountLine()),
		sql.Named("MinBet", side.GetCheck().GetMinBet()),
		sql.Named("MaxBet", side.GetCheck().GetMaxBet()),
		sql.Named("Balance", side.GetCheck().GetBalance()),
		sql.Named("Currency", side.GetCheck().GetCurrency()),
		sql.Named("CheckDone", side.GetCheck().GetDone()),

		sql.Named("CalcStatus", side.GetCheckCalc().GetStatus()),
		sql.Named("MaxStake", side.GetCheckCalc().GetMaxStake()),
		sql.Named("MinStake", side.GetCheckCalc().GetMinStake()),
		sql.Named("MaxWin", side.GetCheckCalc().GetMaxWin()),
		sql.Named("Stake", side.GetCheckCalc().GetStake()),
		sql.Named("Win", side.GetCheckCalc().GetWin()),
		sql.Named("IsFirst", side.GetCheckCalc().GetIsFirst()),
	)
	if err != nil {
		return errors.Wrapf(err, "uspSaveSide error")
	}
	return nil
}

type DemoEvent struct {
	EventId    string
	SportName  string
	SportId    int64
	LeagueName string
	LeagueId   int64
	Home       string
	Away       string
	EventState string
}

func (s *Store) GetDemoEvents(ctx context.Context, count int, sportId int64, homeLike string) (events []DemoEvent, err error) {
	err = s.db.SelectContext(ctx, &events, "select EventId, SportName, SportId, LeagueName, LeagueId, Home, Away, EventState from dbo.fnGetDemoEvents(@p1, @p2, @p3)", count, sportId, homeLike)
	return
}
func (s *Store) FindTeamByName(ctx context.Context, teamName string, sportId int64) (id int64, err error) {
	err = s.db.GetContext(ctx, &id, "select Id from dbo.Team where Name=@p1 and SportId=@p2", teamName, sportId)
	return
}

func (s *Store) FindLineByBetTypeId(ctx context.Context, betTypeId int64, eventId int64) (lines []Ticket, err error) {
	err = s.db.SelectContext(ctx, &lines, "select Id, BetTypeId, EventId,Points from dbo.Line where BetTypeId=@p1 and EventId=@p2", betTypeId, eventId)
	return
}

func (s *Store) SaveBet(sb *pb.Surebet) error {
	side := sb.Members[0]
	_, err := s.db.Exec("dbo.uspSaveBet",
		sql.Named("SurebetId", sb.SurebetId),
		sql.Named("SideIndex", side.Num-1),

		sql.Named("BetId", side.ToBet.Id),
		sql.Named("TryCount", side.GetToBet().GetTryCount()),
		sql.Named("Status", side.GetBet().GetStatus()),
		sql.Named("StatusInfo", side.GetBet().GetStatusInfo()),
		sql.Named("Start", side.GetBet().GetStart()),
		sql.Named("Done", side.GetBet().GetDone()),
		sql.Named("Price", side.GetBet().GetPrice()),
		sql.Named("Stake", side.GetBet().GetStake()),
		sql.Named("ApiBetId", side.GetBet().GetApiBetId()),
	)
	if err != nil {
		return errors.Wrap(err, "uspSaveBet error")
	}
	return nil
}

func (s *Store) Demo() error {
	_, err := s.db.Exec("insert into dbo.Sport (Id) values (999)")
	if err != nil {
		n := err.(mssql.Error)
		//err := errors.Unwrap(err)
		s.log.Error(n)
	}
	return nil

}

func (s *Store) GetStartTime(ctx context.Context, side *pb.SurebetSide) (err error) {
	key := fmt.Sprintf("starts:%v", side.EventId)
	got, b := s.Cache.Get(key)
	if b {
		side.Starts = got.(string)
	}
	err = s.db.GetContext(ctx, &side.Starts, "select Starts from dbo.Event where Id=@p1", side.EventId)
	if err != nil {
		return err
	}
	s.Cache.SetWithTTL(key, side.Starts, 1, time.Minute)
	return
}

const deleteLineQ = `
delete
from dbo.Line
where Id=@p1
`

func (s *Store) DeleteLine(ctx context.Context, lineId int64) error {
	_, err := s.db.ExecContext(ctx, deleteLineQ, lineId)
	if err != nil {
		return err
	}
	return nil
}
