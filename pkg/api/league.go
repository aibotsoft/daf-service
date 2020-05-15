package api

import (
	"context"
	"github.com/antchfx/htmlquery"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"net/url"
	"strconv"
)

var leaguePath, _ = url.Parse(`League.aspx`)

type League struct {
	Id         int64
	Name       string
	SportId    int64
	EventState string
}

func (a *Api) GetLeagues(ctx context.Context, base *url.URL, eventState string, sportId int64) ([]League, error) {
	var leagues []League
	resp, err := a.client.R().EnableTrace().SetContext(ctx).SetDoNotParseResponse(true).
		SetQueryParam("eventState", eventState).
		SetQueryParam("sport", strconv.FormatInt(sportId, 10)).
		Get(base.ResolveReference(leaguePath).String())
	if err != nil {
		return nil, errors.Wrapf(err, "get leagues error for sport %v", sportId)
	}
	page, err := parse(resp)
	if err != nil {
		return nil, err
	}
	links, err := htmlquery.QueryAll(page, `//a[contains(@href,'match.aspx')]`)
	if err != nil {
		return nil, errors.Wrapf(err, "query //a in league page error for sport %v", sportId)
	}
	for _, link := range links {
		leagueId, err := FindLeagueId(link)
		if err != nil {
			a.log.Warn(err)
			continue
		}
		leagues = append(leagues, League{Id: leagueId, Name: InnerTextTrimmed(link), SportId: sportId, EventState: eventState})
	}
	return leagues, nil
}

func FindLeagueId(node *html.Node) (int64, error) {
	u, err := url.Parse(htmlquery.SelectAttr(node, "href"))
	if err != nil {
		return 0, err
	}
	q := u.Query()
	leagueID := q["leagueID"]
	if len(leagueID) == 0 {
		return 0, errors.Errorf("not found leagueId")
	}
	id, err := strconv.ParseInt(leagueID[0], 10, 64)
	if err != nil {
		return 0, errors.Errorf("not found leagueId")
	}
	return id, nil
}
