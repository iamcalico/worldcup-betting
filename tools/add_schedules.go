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
	{
		"俄罗斯", "沙特阿拉伯",
		1.233, 2.345, 2.345,
		"2018-06-14 23:00:00", "A", 0, false,
	},
	{
		"埃及", "乌拉圭",
		1.233, 2.345, 2.345,
		"2018-06-15 20:00:00", "A", 0, false,
	},
	{
		"摩洛哥", "伊朗",
		1.233, 2.345, 2.345,
		"2018-06-15 23:00:00", "B", 0, false,
	},
	{
		"葡萄牙", "西班牙",
		1.233, 2.345, 2.345,
		"2018-06-16 02:00:00", "B", 0, false,
	},
	{
		"法国", "澳大利亚",
		1.233, 2.345, 2.345,
		"2018-06-16 18:00:00", "C", 0, false,
	},
	{
		"阿根廷", "冰岛",
		1.233, 2.345, 2.345,
		"2018-06-16 21:00:00", "D", 0, false,
	},
	{
		"秘鲁", "丹麦",
		1.233, 2.345, 2.345,
		"2018-06-17 00:00:00", "C", 0, false,
	},
	{
		"克罗地亚", "尼日利亚",
		1.233, 2.345, 2.345,
		"2018-06-17 03:00:00", "D", 0, false,
	},
	{
		"哥斯达黎加", "塞尔利亚",
		1.233, 2.345, 2.345,
		"2018-06-17 20:00:00", "E", 0, false,
	},
	{
		"德国", "墨西哥",
		1.233, 2.345, 2.345,
		"2018-06-17 23:00:00", "F", 0, false,
	},
	{
		"巴西", "瑞士",
		1.233, 2.345, 2.345,
		"2018-06-18 02:00:00", "E", 0, false,
	},
	{
		"瑞典", "韩国",
		1.233, 2.345, 2.345,
		"2018-06-18 20:00:00", "F", 0, false,
	},
	{
		"比利时", "巴拿马",
		1.233, 2.345, 2.345,
		"2018-06-18 23:00:00", "G", 0, false,
	},
	{
		"突尼斯", "英格兰",
		1.233, 2.345, 2.345,
		"2018-06-19 02:00:00", "G", 0, false,
	},
	{
		"哥伦比亚", "日本",
		1.233, 2.345, 2.345,
		"2018-06-19 20:00:00", "H", 0, false,
	},
	{
		"波兰", "塞内加尔",
		1.233, 2.345, 2.345,
		"2018-06-19 23:00:00", "H", 0, false,
	},
	{
		"俄罗斯", "埃及",
		1.233, 2.345, 2.345,
		"2018-06-20 02:00:00", "A", 0, false,
	},
	{
		"葡萄牙", "摩洛哥",
		1.233, 2.345, 2.345,
		"2018-06-20 20:00:00", "B", 0, false,
	},
	{
		"乌拉圭", "沙特阿拉伯",
		1.233, 2.345, 2.345,
		"2018-06-20 23:00:00", "A", 0, false,
	},
	{
		"伊朗", "西班牙",
		1.233, 2.345, 2.345,
		"2018-06-21 02:00:00", "B", 0, false,
	},
	{
		"丹麦", "澳大利亚",
		1.233, 2.345, 2.345,
		"2018-06-21 20:00:00", "C", 0, false,
	},
	{
		"法国", "秘鲁",
		1.233, 2.345, 2.345,
		"2018-06-21 23:00:00", "C", 0, false,
	},
	{
		"阿根廷", "克罗地亚",
		1.233, 2.345, 2.345,
		"2018-06-22 02:00:00", "D", 0, false,
	},
	{
		"巴西", "哥斯达黎加",
		1.233, 2.345, 2.345,
		"2018-06-22 20:00:00", "E", 0, false,
	},
	{
		"尼日利亚", "冰岛",
		1.233, 2.345, 2.345,
		"2018-06-22 23:00:00", "D", 0, false,
	},
	{
		"塞尔维亚", "瑞士",
		1.233, 2.345, 2.345,
		"2018-06-23 02:00:00", "E", 0, false,
	},
	{
		"比利时", "突尼斯",
		1.233, 2.345, 2.345,
		"2018-06-23 20:00:00", "G", 0, false,
	},
	{
		"韩国", "墨西哥",
		1.233, 2.345, 2.345,
		"2018-06-23 23:00:00", "F", 0, false,
	},
	{
		"德国", "瑞典",
		1.233, 2.345, 2.345,
		"2018-06-24 02:00:00", "F", 0, false,
	},
	{
		"英格兰", "巴拿马",
		1.233, 2.345, 2.345,
		"2018-06-24 20:00:00", "G", 0, false,
	},
	{
		"日本", "塞内加尔",
		1.233, 2.345, 2.345,
		"2018-06-24 23:00:00", "H", 0, false,
	},
}

func main() {
	for _, schedule := range schedules {
		jsonData, err := json.Marshal(schedule)
		if err != nil {
			log.Fatalf("json marshal error: %v\n", err)
		}
		req, err := http.NewRequest("PUT", "http://z1.zhengyinyong.com:9614/new_schedule", bytes.NewBuffer(jsonData))
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
