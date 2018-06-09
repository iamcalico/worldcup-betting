package main

import (
	"bytes"
	"log"
	"net/http"

	"io/ioutil"

	"github.com/gin-gonic/gin/json"
)

type ScheduleType int

const (
	GroupMatches       ScheduleType = 0 // 小组赛
	RoundEight                          // 八强赛
	FinalFour                           // 四强赛
	Semifinal                           // 半决赛
	MatchForThirdPlace                  // 季军赛
	Finals                              // 总决赛
)

type ScheduleStatus int

const (
	HomeTeamWin ScheduleStatus = 0 // 主队胜利
	AwayTeamWin                    // 客队胜利
	Draw                           // 平局
)

type NewScheduleReq struct {
	HomeTeam        string       `json:"home_team"`          // 主队
	AwayTeam        string       `json:"away_team"`          // 客队
	HomeTeamWinOdds float64      `json:"home_team_win_odds"` // 主队胜利的赔率
	AwayTeamWinOdds float64      `json:"away_team_win_odds"` // 客队胜利的赔率
	TiedOdds        float64      `json:"tied_odds"`          // 平局的赔率
	ScheduleTime    string       `json:"schedule_time"`      // 比赛时间
	ScheduleGroup   string       `json:"schedule_group"`     // 比赛组别
	ScheduleType    ScheduleType `json:"schedule_type"`      // 比赛类别
	DisableBetting  bool         `json:"disable_betting"`    // 是否允许投注
}

var schedules = []NewScheduleReq{
	{"Russia", "Saudi Arabia", 1.233, 2.345, 2.345, "2018-06-14 23:00:00", "A", 0, true},
}

func main() {
	for _, schedule := range schedules {
		jsonData, err := json.Marshal(schedule)
		if err != nil {
			log.Fatalf("json marshal error: %v\n", err)
		}
		req, err := http.NewRequest("PUT", "http://localhost:9614/new_schedule", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("do post request failed, error: %v, body: %v\n", err, body)
		}
	}
}
