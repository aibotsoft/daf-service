package store

import (
	"context"
	"database/sql"
	"github.com/aibotsoft/daf-service/pkg/api"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/cache"
	"github.com/aibotsoft/micro/config"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/dgraph-io/ristretto"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strconv"
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

func (s *Store) LoadToken(ctx context.Context) (*api.Token, error) {
	s.log.Info("begin load token from db")
	var t api.Token
	err := s.db.GetContext(ctx, &t, "select top 1 Id, Host, Token,  LastCheckAt Last from dbo.Auth order by LastCheckAt desc")
	if err != nil {
		return nil, errors.Wrap(err, "load token error")
	}
	//a.log.Infow("loaded token", "host", a.host, "token", a.token, "last", a.last, "id", a.id)
	return &t, nil
}

var sportCache = make(map[int64]api.Sport)

func (s *Store) SaveSports(ctx context.Context, sports []api.Sport) error {
	var newSports []api.Sport
	for _, sport := range sports {
		_, ok := sportCache[sport.Id]
		if !ok {
			newSports = append(newSports, sport)
			sportCache[sport.Id] = sport
		}
	}
	if len(newSports) == 0 {
		return nil
	}
	tvp := mssql.TVP{TypeName: "SportType", Value: newSports}
	_, err := s.db.ExecContext(ctx, "dbo.uspCreateSports", tvp)
	if err != nil {
		return errors.Wrapf(err, "uspCreateSport error for %v", newSports)
	}
	return nil
}

var leagueCache = make(map[int64]api.League)

func (s *Store) SaveLeagues(ctx context.Context, leagues []api.League) error {
	var newLeagues []api.League
	for _, league := range leagues {
		_, ok := leagueCache[league.Id]
		if !ok {
			newLeagues = append(newLeagues, league)
			leagueCache[league.Id] = league
		}
	}
	if len(newLeagues) == 0 {
		return nil
	}
	tvp := mssql.TVP{TypeName: "LeagueType", Value: newLeagues}
	_, err := s.db.ExecContext(ctx, "dbo.uspCreateLeagues", tvp)
	if err != nil {
		return errors.Wrapf(err, "uspCreateLeagues error for %v", newLeagues)
	}
	return nil
}

var eventCache = make(map[int64]api.Event)

func (s *Store) SaveEvents(ctx context.Context, events []api.Event) error {
	var newEvents []api.Event
	for i := range events {
		got, ok := eventCache[events[i].Id]
		if !ok {
			newEvents = append(newEvents, events[i])
			eventCache[events[i].Id] = events[i]
		} else {
			s.log.Infow("double_event", "old", got, "new", events[i])
		}
	}
	if len(newEvents) == 0 {
		return nil
	}
	tvp := mssql.TVP{TypeName: "EventType", Value: newEvents}
	_, err := s.db.ExecContext(ctx, "dbo.uspCreateEvents", tvp)
	if err != nil {
		return errors.Wrapf(err, "uspCreateEvents error for %v", newEvents)
	}
	return nil
}

type lineKey struct {
	id      int64
	betTeam string
}

func (s *Store) SaveMarkets(ctx context.Context, id int64, name string) error {
	key := "market:" + strconv.FormatInt(id, 10)
	_, b := s.Cache.Get(key)
	if b {
		return nil
	}
	_, err := s.db.ExecContext(ctx, "dbo.uspCreateMarket", id, name)
	s.Cache.Set(key, id, 1)
	return err
}
func (s *Store) SaveLines(ctx context.Context, tickets []api.Ticket) error {
	var newTickets []api.Ticket
	var ticketCache = make(map[lineKey]api.Ticket)

	for _, ticket := range tickets {
		err := s.SaveMarkets(ctx, ticket.BetTypeId, ticket.MarketName)
		if err != nil {
			s.log.Errorw("save market error")
		}
		key := lineKey{id: ticket.Id, betTeam: ticket.BetTeam}
		got, ok := ticketCache[key]
		if !ok {
			newTickets = append(newTickets, ticket)
			ticketCache[key] = ticket
		} else {
			s.log.Infow("ticket double", "first", got, "second", ticket)
		}
	}
	if len(newTickets) == 0 {
		return nil
	}
	tvp := mssql.TVP{TypeName: "LineType", Value: newTickets}
	_, err := s.db.ExecContext(ctx, "dbo.uspCreateLines", tvp)
	if err != nil {
		return errors.Wrapf(err, "uspCreateLines error for %v", newTickets)
	}
	return nil
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
select Id from dbo.Event where Home = @p1 and Away=@p2 and LeagueId=@p3 and SportId=@p4
`

func (s *Store) FindEventByName(ctx context.Context, home string, away string, leagueId int64, sportId int64) (int64, error) {
	var id int64
	err := s.db.GetContext(ctx, &id, findEventQ, home, away, leagueId, sportId)
	return id, err
}

//func (s *Store) FindEvent(ctx context.Context, side *pb.SurebetSide) error {
//	key := "event:" + strconv.FormatInt(side.Forted.EventId, 10)
//	got, b := s.Cache.Get(key)
//	if b {
//		side.EventId = got.(int64)
//		return nil
//	}
//	var err error
//	side.SportId, err = s.FindSportByName(ctx, side.SportName)
//	if err != nil {
//		return errors.Wrapf(err, "not found sport by name: %q in db", side.SportName)
//	}
//	side.LeagueId, err = s.FindLeagueByName(ctx, side.SportId, side.LeagueName)
//	if err != nil {
//		return errors.Wrapf(err, "not found league in db, name: %q, sportId: %v", side.LeagueName, side.SportId)
//	}
//	side.EventId, err = s.FindEventByName(ctx, side.Home, side.Away, side.LeagueId, side.SportId)
//	if err != nil {
//		return errors.Wrapf(err, "not found event in db, home: %q, away: %q, leagueId: %d, sportId: %d", side.Home, side.Away, side.LeagueId, side.SportId)
//	}
//	s.Cache.Set(key, side.EventId, 1)
//	return nil
//}

//const findLineQ = "select Id from dbo.Line where EventId = @p1 and BetTeam = @p2 and BetTypeId = @p3"

func (s *Store) FindLine(ctx context.Context, line *api.Ticket) error {
	//err := s.db.GetContext(ctx, &line.Id, findLineQ, side.EventId, line.BetTeam, line.BetTypeId)
	err := s.db.GetContext(ctx, &line.Id, "dbo.uspFindLine",
		sql.Named("BetTeam", line.BetTeam),
		sql.Named("BetTypeId", line.BetTypeId),
		sql.Named("EventId", line.EventId),
		sql.Named("Points", line.Points),
	)
	return err
}

func (s *Store) GetStat(side *pb.SurebetSide) error {
	err := s.db.Get(side.Check, "dbo.uspCalcStat", sql.Named("EventId", side.EventId), sql.Named("MarketName", side.MarketName))
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "uspCalcStat error")
	}
	return nil
}

func (s *Store) SaveCheck(sb *pb.Surebet) error {
	side := sb.Members[0]
	_, err := s.db.Exec("dbo.uspSaveSide",
		sql.Named("Id", sb.SurebetId),
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
