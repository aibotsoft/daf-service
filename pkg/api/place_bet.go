package api

import (
	"context"
	"github.com/aibotsoft/micro/status"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var placeBetPath, _ = url.Parse(`underover/BetNew.aspx`)
var betIdRe = regexp.MustCompile(`ID:\s*(\d+)\s*<`)

var bodyRe = regexp.MustCompile(`(?s)<body>(.*?)</body>`)

func parseBetId(resp string) (int, error) {
	found := betIdRe.FindStringSubmatch(resp)
	if len(found) < 2 {
		return 0, errors.Errorf("not found bet id in %s", resp)
	}
	betId, err := strconv.Atoi(found[1])
	if err != nil {
		return 0, errors.Wrapf(err, "cant parse bet id in %s", found[1])
	}
	return betId, nil
}

type PlaceBetResult struct {
	Status     string
	StatusInfo string
	BetId      int
	Body       string
}

func GetBody(resp string) string {
	found := bodyRe.FindStringSubmatch(resp)
	if len(found) == 2 {
		return found[1]
	}
	return resp
}

const (
	oddsChange    = "Odds has changed"
	oddsSuspended = "Odds suspended"
	EventClosed   = "Event Closed"
)

func (a *Api) PlaceBet(ctx context.Context, base *url.URL, line *Ticket, stake float64) (PlaceBetResult, error) {
	a.client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(20))
	defer a.client.SetRedirectPolicy(resty.NoRedirectPolicy())

	betResult := PlaceBetResult{Status: status.StatusNotAccepted}

	resp, err := a.client.R().EnableTrace().SetContext(ctx).
		SetQueryParam("BetType", strconv.FormatInt(line.BetTypeId, 10)).
		SetQueryParam("oddsid", strconv.FormatInt(line.Id, 10)).
		SetQueryParam("betteam", line.BetTeam).
		SetFormData(map[string]string{"stake": "1", "cmdBetterOddsAndBet": "Accept Better Odds and Bet"}).
		Post(base.ResolveReference(placeBetPath).String())

	if err != nil {
		return betResult, errors.Wrapf(err, "place bet error")
	}
	body := resp.String()

	id, err := parseBetId(body)
	if err == nil {
		return PlaceBetResult{Status: status.StatusOk, BetId: id}, nil
	}
	//found := strings.Index(resp.String(), "General Error")

	if strings.Index(body, EventClosed) != -1 {
		betResult.StatusInfo = EventClosed

	} else if strings.Index(body, oddsChange) != -1 {
		betResult.StatusInfo = oddsChange

	} else if strings.Index(body, oddsSuspended) != -1 {
		betResult.StatusInfo = oddsSuspended

	} else {
		betResult.Body = GetBody(body)

	}
	return betResult, nil
}
