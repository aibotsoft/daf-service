package store

import (
	"context"
	pb "github.com/aibotsoft/gen/fortedpb"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/pkg/errors"
	"time"
)

type Bet struct {
	Id         int64
	SportName  string
	SportId    int64
	LeagueName string
	LeagueId   int64
	Home       string
	Away       string
	EventId    int64
	MarketName string
	SideName   string

	Points      *float64
	EventTime   *time.Time
	Status      string
	WinLoss     *float64
	Price       float64
	Stake       float64
	BetTypeId   int64
	BetTime     *time.Time
	WinLoseDate *time.Time
}

func (s *Store) GetResults(ctx context.Context) ([]pb.BetResult, error) {
	var res []pb.BetResult
	rows, err := s.db.QueryxContext(ctx, "select * from dbo.GetResults ")
	if err != nil {
		return nil, errors.Wrap(err, "get_bet_results_error")
	}
	for rows.Next() {
		var r pb.BetResult
		var Price, Stake, WinLoss *float64
		var ApiBetId, ApiBetStatus *string
		err := rows.Scan(&r.SurebetId, &r.SideIndex, &r.BetId, &ApiBetId, &ApiBetStatus, &Price, &Stake, &WinLoss)
		if err != nil {
			s.log.Error(err)
			continue
		}
		if ApiBetId != nil {
			r.ApiBetId = *ApiBetId
		}
		if ApiBetStatus != nil {
			r.ApiBetStatus = *ApiBetStatus
		}
		if Price != nil {
			r.Price = *Price
		}
		if Stake != nil {
			r.Stake = *Stake
		}
		if WinLoss != nil {
			r.WinLoss = *WinLoss
		}
		//s.log.Infow("res", "", r)
		res = append(res, r)
	}
	return res, nil
}

func (s *Store) SaveBetList(ctx context.Context, bets []Bet) error {
	if len(bets) == 0 {
		return nil
	}
	tvp := mssql.TVP{TypeName: "BetListType", Value: bets}
	_, err := s.db.ExecContext(ctx, "dbo.uspCreateBetList", tvp)
	if err != nil {
		return errors.Wrap(err, "uspCreateBetList_error")
	}
	return nil
}
