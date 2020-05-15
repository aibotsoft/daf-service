package api

import (
	"context"
	"github.com/pkg/errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var ticketPath, _ = url.Parse(`underover/BetNew.aspx`)
var teamsRe = regexp.MustCompile(`<b>\s?(.*?)\s*-vs-\s*(.*?)\s*(</b>)?<br>`)
var minBetRe = regexp.MustCompile(`Min:\s*(\d*)\s*<br>`)
var maxBetRe = regexp.MustCompile(`Max:\s*(\d*)\s*<br>`)
var marketPriceRe = regexp.MustCompile(`(.*)\s?@\s?(.*)\s?\[Dec\]`)

var ticketBodyRe = regexp.MustCompile(`(?s)<b>(.*)</b>`)

func parseFloat(re *regexp.Regexp, ticket string) (float64, error) {
	vStr := re.FindStringSubmatch(ticket)
	if len(vStr) < 2 {
		return 0, errors.Errorf("error find value from ticket %v", ticket)
	}
	valueNoComma := strings.Replace(vStr[1], ",", "", -1)
	value, err := strconv.ParseFloat(valueNoComma, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "error parse float from value %q", valueNoComma)
	}
	return value, nil
}

func (a *Api) GetTicket(ctx context.Context, base *url.URL, line *Ticket) error {
	resp, err := a.client.R().EnableTrace().SetContext(ctx).
		SetQueryParam("oddsid", strconv.FormatInt(line.Id, 10)).
		SetQueryParam("betteam", line.BetTeam).
		SetQueryParam("BetType", strconv.FormatInt(line.BetTypeId, 10)).
		Get(base.ResolveReference(ticketPath).String())
	if err != nil {
		return errors.Wrapf(err, "get ticket error")
	}
	body := resp.String()
	ticketBody := ticketBodyRe.FindString(body)
	a.log.Info(ticketBody)
	//teams := teamsRe.FindStringSubmatch(ticketBody)
	//if len(teams) < 3 {
	//	return errors.Errorf("error get teams from ticket %v", ticketBody)
	//}
	//line.Home = strings.TrimSpace(teams[1])
	//line.Away = strings.TrimSpace(teams[2])
	line.Home, line.Away, err = parseTeams(ticketBody)
	if err != nil {
		return err
	}

	line.MinBet, err = parseFloat(minBetRe, ticketBody)
	if err != nil {
		return errors.Wrapf(err, "minBet error for %v", ticketBody)
	}
	line.MaxBet, err = parseFloat(maxBetRe, ticketBody)
	if err != nil {
		return errors.Wrapf(err, "maxBet error for %v", ticketBody)
	}
	scoreFound := strings.Index(ticketBody, "Score:")
	if scoreFound != -1 {
		line.IsLive = true
	}
	market := marketPriceRe.FindStringSubmatch(ticketBody)
	a.log.Infow("market", "m", market)
	if len(market) < 3 {
		return errors.Errorf("error get market from ticket %v", ticketBody)
	}
	points, err := FindPointsInName(market[1])
	if err != nil {
		return err
	}
	if points != nil && *points != *line.Points {
		a.log.Infow("points", "should", *line.Points, "actual", *points)
		return DiffPointsError
	}
	line.Price, err = strconv.ParseFloat(market[2], 64)
	if err != nil {
		return errors.Wrapf(err, "error parse price from %q", market[3])
	} else if line.Price == 0 {
		a.log.Info("price == 0")
	}
	a.log.Info(line)
	return nil
}

var DiffPointsError = errors.New("points different")

func FindPointsInName(name string) (*float64, error) {
	foundPoints := pointRe.FindStringSubmatch(name)
	if len(foundPoints) > 1 {
		float, err := strconv.ParseFloat(foundPoints[1], 64)
		if err != nil {
			return nil, errors.Wrapf(err, "find points error in %s", name)
		}
		return &float, nil
	}
	return nil, nil
}
