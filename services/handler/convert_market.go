package handler

import (
	"github.com/aibotsoft/daf-service/pkg/api"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
)

var handicapRe = regexp.MustCompile(`Ф(\d)\((-?\d+,?\d{0,3})\)`)
var totalRe = regexp.MustCompile(`Т([МБ])\((-?\d+,?\d{0,3})\)`)
var teamTotalRe = regexp.MustCompile(`ИТ(\d)([МБ])\((-?\d+,?\d{0,3})\)`)
var handicapSide = map[string]string{"1": "h", "2": "a"}
var totalSide = map[string]string{"Б": "h", "М": "a"}

const (
	Handicap          = 1
	OddEven           = 2
	OverUnder         = 3
	FT1X2             = 5
	FHHandicap        = 7
	FHOverUnder       = 8
	FHOddEven         = 12
	FH1X2             = 15
	Moneyline         = 20
	Quarter1Handicap  = 609
	Quarter1OverUnder = 610
	Quarter1OddEven   = 611
)

var abc = map[string]api.Ticket{
	"П1": {BetTypeId: Moneyline, BetTeam: "h"},
	"П2": {BetTypeId: Moneyline, BetTeam: "a"},

	"(сеты) П1": {BetTypeId: Moneyline, BetTeam: "h"},
	"(сеты) П2": {BetTypeId: Moneyline, BetTeam: "a"},

	//"1/2 П1": {MarketName: "First Half ML - Money Line", BetTeam: "1"},
	//"1/2 П2": {MarketName: "First Half ML - Money Line", BetTeam: "2"},
	//"1/4 П1": {MarketName: "1st Quarter - Money Line", BetTeam: "1"},
	//"1/4 П2": {MarketName: "1st Quarter - Money Line", BetTeam: "2"},
	//"2/4 П1": {MarketName: "2nd Quarter - Money Line", BetTeam: "1"},
	//"2/4 П2": {MarketName: "2nd Quarter - Money Line", BetTeam: "2"},
	//"3/4 П1": {MarketName: "3rd Quarter - Money Line", BetTeam: "1"},
	//"3/4 П2": {MarketName: "3rd Quarter - Money Line", BetTeam: "2"},
	//"4/4 П1": {MarketName: "4th Quarter - Money Line", BetTeam: "1"},
	//"4/4 П2": {MarketName: "4th Quarter - Money Line", BetTeam: "2"},

	"Чёт":       {BetTypeId: OddEven, BetTeam: "a"},
	"Нечёт":     {BetTypeId: OddEven, BetTeam: "h"},
	"1/2 Чёт":   {BetTypeId: FHOddEven, BetTeam: "a"},
	"1/2 Нечёт": {BetTypeId: FHOddEven, BetTeam: "h"},
	"1/4 Чёт":   {BetTypeId: Quarter1OddEven, BetTeam: "a"},
	"1/4 Нечёт": {BetTypeId: Quarter1OddEven, BetTeam: "h"},
	//"2/4 Чёт":   {BetTypeId: "2nd Quarter - Odd/Even", BetTeam: "a"},
	//"2/4 Нечёт": {BetTypeId: "2nd Quarter - Odd/Even", BetTeam: "h"},
	//"3/4 Чёт":   {BetTypeId: "3rd Quarter - Odd/Even", BetTeam: "a"},
	//"3/4 Нечёт": {BetTypeId: "3rd Quarter - Odd/Even", BetTeam: "h"},
	//"4/4 Чёт":   {BetTypeId: "4th Quarter - Odd/Even", BetTeam: "a"},
	//"4/4 Нечёт": {BetTypeId: "4th Quarter - Odd/Even", BetTeam: "h"},

	"1":     {BetTypeId: FT1X2, BetTeam: "1"},
	"2":     {BetTypeId: FT1X2, BetTeam: "2"},
	"Х":     {BetTypeId: FT1X2, BetTeam: "x"},
	"1/2 1": {BetTypeId: FH1X2, BetTeam: "1"},
	"1/2 2": {BetTypeId: FH1X2, BetTeam: "2"},
	"1/2 Х": {BetTypeId: FH1X2, BetTeam: "x"},

	"УГЛ 1":     {BetTypeId: FT1X2, BetTeam: "1"},
	"УГЛ 2":     {BetTypeId: FT1X2, BetTeam: "2"},
	"УГЛ Х":     {BetTypeId: FT1X2, BetTeam: "x"},
	"УГЛ 1/2 1": {BetTypeId: FH1X2, BetTeam: "1"},
	"УГЛ 1/2 2": {BetTypeId: FH1X2, BetTeam: "2"},
	"УГЛ 1/2 Х": {BetTypeId: FH1X2, BetTeam: "x"},

	"ЖК 1":     {BetTypeId: FT1X2, BetTeam: "1"},
	"ЖК 2":     {BetTypeId: FT1X2, BetTeam: "2"},
	"ЖК Х":     {BetTypeId: FT1X2, BetTeam: "x"},
	"ЖК 1/2 1": {BetTypeId: FH1X2, BetTeam: "1"},
	"ЖК 1/2 2": {BetTypeId: FH1X2, BetTeam: "2"},
	"ЖК 1/2 Х": {BetTypeId: FH1X2, BetTeam: "x"},

	//"1X": {MarketName: "Double Chance", BetTeam: "1x"},
	//"X2": {MarketName: "Double Chance", BetTeam: "x2"},
	//"12": {MarketName: "Double Chance", BetTeam: "12"},
}

func processPoint(point string) (float64, error) {
	pointsStr := strings.Replace(point, ",", ".", -1)
	return strconv.ParseFloat(pointsStr, 64)
}

func Convert(name string) (*api.Ticket, error) {
	if name == "" {
		return nil, errors.Errorf("cannot convert empty str")
	}
	m := abc[name]
	if m.BetTypeId != 0 {
		return &m, nil
	}
	var found []string
	found = handicapRe.FindStringSubmatch(name)
	if len(found) == 3 {
		return processHandicap(name, found)
	}
	found = totalRe.FindStringSubmatch(name)
	if len(found) == 3 {
		return processTotal(name, found)
	}
	found = teamTotalRe.FindStringSubmatch(name)
	if len(found) == 4 {
		return processTeamTotal(name, found)
	}
	return nil, errors.Errorf("not converted market: %s", name)
}
func processHandicap(market string, found []string) (*api.Ticket, error) {
	point, err := processPoint(found[2])
	if err != nil {
		return nil, err
	}
	typeId := Handicap
	switch {
	case strings.Index(market, "1/2") != -1:
		typeId = FHHandicap
	case strings.Index(market, "1/4") != -1:
		typeId = Quarter1Handicap
		//case strings.Index(market, "2/4") != -1:
		//	typeId = "2nd Quarter - Handicap"
		//case strings.Index(market, "3/4") != -1:
		//	typeId = "3rd Quarter - Handicap"
		//case strings.Index(market, "4/4") != -1:
		//	typeId = "4th Quarter - Handicap"
	}
	m := api.Ticket{BetTypeId: int64(typeId), BetTeam: handicapSide[found[1]], Points: &point}
	return &m, nil
}

func processTotal(market string, found []string) (*api.Ticket, error) {
	point, err := processPoint(found[2])
	if err != nil {
		return nil, err
	}
	typeId := OverUnder
	switch {
	case strings.Index(market, "1/2") != -1:
		typeId = FHOverUnder
	case strings.Index(market, "1/4") != -1:
		typeId = Quarter1OverUnder
		//case strings.Index(market, "2/4") != -1:
		//	typeId = "2nd Quarter - Over/Under"
		//case strings.Index(market, "3/4") != -1:
		//	typeId = "3rd Quarter - Over/Under"
		//case strings.Index(market, "4/4") != -1:
		//	typeId = "4th Quarter - Over/Under"
	}
	m := api.Ticket{BetTypeId: int64(typeId), BetTeam: totalSide[found[1]], Points: &point}
	return &m, nil
}
func processTeamTotal(market string, found []string) (*api.Ticket, error) {
	point, err := processPoint(found[3])
	if err != nil {
		return nil, err
	}
	name := "Away Team Total Points - Over/Under"
	if found[1] == "1" {
		name = "Home Team Total Points - Over/Under"
	}
	m := api.Ticket{MarketName: name, BetTeam: totalSide[found[2]], Points: &point}
	return &m, nil
}
