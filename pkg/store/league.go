package store

import (
	"context"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/pkg/errors"
	"sync"
)

type League struct {
	Id         int64
	Name       string
	SportId    int64
	EventState string
}

func (t *League) Equal(that interface{}) bool {
	if that == nil {
		return t == nil
	}
	that1, ok := that.(*League)
	if !ok {
		that2, ok := that.(League)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return t == nil
	} else if t == nil {
		return false
	}
	if t.Name != that1.Name {
		return false
	}
	if t.SportId != that1.SportId {
		return false
	}
	return true
}

var leagueCache sync.Map

func (s *Store) SaveLeagues(ctx context.Context, leagues []League) error {
	var newLeagues []League
	for i := range leagues {
		got, ok := leagueCache.Load(leagues[i].Id)
		if !ok {
			newLeagues = append(newLeagues, leagues[i])
			leagueCache.Store(leagues[i].Id, leagues[i])
		} else if !leagues[i].Equal(got) {
			newLeagues = append(newLeagues, leagues[i])
			leagueCache.Store(leagues[i].Id, leagues[i])
		}
	}
	if len(newLeagues) == 0 {
		//s.log.Debug("no newLeagues")
		return nil
	}
	s.log.Debugw("got_new_leagues", "count", len(newLeagues))

	tvp := mssql.TVP{TypeName: "LeagueType", Value: newLeagues}
	_, err := s.db.ExecContext(ctx, "dbo.uspCreateLeagues", tvp)
	if err != nil {
		return errors.Wrapf(err, "uspCreateLeagues error for %v", newLeagues)
	}
	return nil
}
