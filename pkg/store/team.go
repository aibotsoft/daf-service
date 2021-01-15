package store

import (
	"context"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/pkg/errors"
	"sync"
)

type Team struct {
	Id      int64
	Name    string
	SportId int64
}

func (t *Team) Equal(that interface{}) bool {
	if that == nil {
		return t == nil
	}
	that1, ok := that.(*Team)
	if !ok {
		that2, ok := that.(Team)
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

var teamCache sync.Map

func (s *Store) SaveTeams(ctx context.Context, teams []Team) error {
	var newTeams []Team
	for i := range teams {
		got, ok := teamCache.Load(teams[i].Id)
		if !ok {
			newTeams = append(newTeams, teams[i])
			teamCache.Store(teams[i].Id, teams[i])
		} else if !teams[i].Equal(got) {
			newTeams = append(newTeams, teams[i])
			teamCache.Store(teams[i].Id, teams[i])
		}
	}
	if len(newTeams) == 0 {
		//s.log.Debug("no newTeams")
		return nil
	}
	s.log.Debugw("got_new_teams", "count", len(newTeams))

	tvp := mssql.TVP{TypeName: "TeamType", Value: newTeams}
	_, err := s.db.ExecContext(ctx, "dbo.uspCreateTeams", tvp)
	if err != nil {
		return errors.Wrapf(err, "uspCreateTeams error for %v", newTeams)
	}
	return nil
}
