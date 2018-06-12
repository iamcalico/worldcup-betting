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
	ScheduleID      int          `json:"schedule_id"`
	HomeTeam        string       `json:"home_team"`          // 主队
	AwayTeam        string       `json:"away_team"`          // 客队
	HomeTeamWinOdds float64      `json:"home_team_win_odds"` // 主队胜利的赔率
	AwayTeamWinOdds float64      `json:"away_team_win_odds"` // 客队胜利的赔率
	TiedOdds        float64      `json:"tied_odds"`          // 平局的赔率
	ScheduleTime    string       `json:"schedule_time"`      // 比赛时间
	ScheduleGroup   string       `json:"schedule_group"`     // 比赛组别
	ScheduleType    ScheduleType `json:"schedule_type"`      // 比赛类别
	ScheduleStatus  int          `json:"schedule_status"`    // 比赛结果
	DisableBetting  bool         `json:"disable_betting"`    // 是否允许投注
	EnableDisplay   bool         `json:"enable_dispaly"`     // 是否显示在投注页
}

const (
	homeTeamWinOdds = 2
	awayTeamWinOdds = 3
	tiedOdds        = 3
)

var schedules = []NewScheduleReq{
	{
		1, "俄罗斯", "沙特阿拉伯",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-01 23:00:00", "A", 0, 0, false, true,
	},
	{
		2, "埃及", "乌拉圭",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-15 20:00:00", "A", 0, 0, false, true,
	},
	{
		3, "摩洛哥", "伊朗",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-15 23:00:00", "B", 0, 0, false, true,
	},
	{
		4, "葡萄牙", "西班牙",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-16 02:00:00", "B", 0, 0, false, true,
	},
	{
		5, "法国", "澳大利亚",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-16 18:00:00", "C", 0, 0, false, true,
	},
	{
		6, "阿根廷", "冰岛",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-16 21:00:00", "D", 0, 0, false, true,
	},
	{
		7, "秘鲁", "丹麦",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-17 00:00:00", "C", 0, 0, false, false,
	},
	{
		8, "克罗地亚", "尼日利亚",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-17 03:00:00", "D", 0, 0, false, false,
	},
	{
		9, "哥斯达黎加", "塞尔维亚",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-17 20:00:00", "E", 0, 0, false, false,
	},
	{
		10, "德国", "墨西哥",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-17 23:00:00", "F", 0, 0, false, false,
	},
	{
		11, "巴西", "瑞士",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-18 02:00:00", "E", 0, 0, false, false,
	},
	{
		12, "瑞典", "韩国",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-18 20:00:00", "F", 0, 0, false, false,
	},
	{
		13, "比利时", "巴拿马",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-18 23:00:00", "G", 0, 0, false, false,
	},
	{
		14, "突尼斯", "英格兰",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-19 02:00:00", "G", 0, 0, false, false,
	},
	{
		15, "哥伦比亚", "日本",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-19 20:00:00", "H", 0, 0, false, false,
	},
	{
		16, "波兰", "塞内加尔",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-19 23:00:00", "H", 0, 0, false, false,
	},
	{
		17, "俄罗斯", "埃及",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-20 02:00:00", "A", 0, 0, false, false,
	},
	{
		18, "葡萄牙", "摩洛哥",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-20 20:00:00", "B", 0, 0, false, false,
	},
	{
		19, "乌拉圭", "沙特阿拉伯",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-20 23:00:00", "A", 0, 0, false, false,
	},
	{
		20, "伊朗", "西班牙",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-21 02:00:00", "B", 0, 0, false, false,
	},
	{
		21, "丹麦", "澳大利亚",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-21 20:00:00", "C", 0, 0, false, false,
	},
	{
		22, "法国", "秘鲁",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-21 23:00:00", "C", 0, 0, false, false,
	},
	{
		23, "阿根廷", "克罗地亚",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-22 02:00:00", "D", 0, 0, false, false,
	},
	{
		24, "巴西", "哥斯达黎加",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-22 20:00:00", "E", 0, 0, false, false,
	},
	{
		25, "尼日利亚", "冰岛",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-22 23:00:00", "D", 0, 0, false, false,
	},
	{
		26, "塞尔维亚", "瑞士",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-23 02:00:00", "E", 0, 0, false, false,
	},
	{
		27, "比利时", "突尼斯",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-23 20:00:00", "G", 0, 0, false, false,
	},
	{
		28, "韩国", "墨西哥",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-23 23:00:00", "F", 0, 0, false, false,
	},
	{
		29, "德国", "瑞典",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-24 02:00:00", "F", 0, 0, false, false,
	},
	{
		30, "英格兰", "巴拿马",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-24 20:00:00", "G", 0, 0, false, false,
	},
	{
		31, "日本", "塞内加尔",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-24 23:00:00", "H", 0, 0, false, false,
	},
	{
		32, "波兰", "哥伦比亚",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-25 02:00:00", "H", 0, 0, false, false,
	},
	{
		33, "沙特阿拉伯", "埃及",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-25 22:00:00", "A", 0, 0, false, false,
	},
	{
		34, "俄罗斯", "乌拉圭",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-25 22:00:00", "A", 0, 0, false, false,
	},
	{
		35, "西班牙", "摩洛哥",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-26 02:00:00", "B", 0, 0, false, false,
	},
	{
		36, "伊朗", "葡萄牙",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-26 02:00:00", "B", 0, 0, false, false,
	},
	{
		37, "丹麦", "法国",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-26 22:00:00", "C", 0, 0, false, false,
	},
	{
		38, "澳大利亚", "秘鲁",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-26 22:00:00", "C", 0, 0, false, false,
	},
	{
		39, "冰岛", "克罗地亚",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-27 02:00:00", "D", 0, 0, false, false,
	},
	{
		40, "尼日利亚", "阿根廷",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-27 02:00:00", "D", 0, 0, false, false,
	},
	{
		41, "墨西哥", "瑞典",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-27 22:00:00", "F", 0, 0, false, false,
	},
	{
		42, "韩国", "德国",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-27 22:00:00", "F", 0, 0, false, false,
	},
	{
		43, "塞尔维亚", "巴西",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-28 02:00:00", "E", 0, 0, false, false,
	},
	{
		44, "瑞士", "哥斯达黎加",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-28 02:00:00", "E", 0, 0, false, false,
	},
	{
		45, "塞内加尔", "哥伦比亚",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-28 22:00:00", "H", 0, 0, false, false,
	},
	{
		46, "日本", "波兰",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-28 22:00:00", "H", 0, 0, false, false,
	},
	{
		47, "英格兰", "比利时",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-29 02:00:00", "G", 0, 0, false, false,
	},
	{
		48, "巴拿马", "突尼斯",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-29 02:00:00", "G", 0, 0, false, false,
	},
	{
		49, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-06-30 22:00:00", "X", 1, 0, true, true,
	},
	{
		50, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-01 02:00:00", "X", 1, 0, true, true,
	},
	{
		51, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-01 22:00:00", "X", 1, 0, true, true,
	},
	{
		52, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-02 02:00:00", "X", 1, 0, true, true,
	},
	{
		53, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-02 22:00:00", "X", 1, 0, true, true,
	},
	{
		54, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-03 02:00:00", "X", 1, 0, true, true,
	},
	{
		55, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-03 22:00:00", "X", 1, 0, true, true,
	},
	{
		56, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-04 02:00:00", "X", 1, 0, true, true,
	},
	{
		57, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-06 22:00:00", "X", 2, 0, true, true,
	},
	{
		58, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-07 02:00:00", "X", 2, 0, true, true,
	},
	{
		59, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-07 22:00:00", "X", 2, 0, true, true,
	},
	{
		60, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-08 02:00:00", "X", 2, 0, true, true,
	},
	{
		61, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-11 02:00:00", "X", 3, 0, true, true,
	},
	{
		62, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-12 02:00:00", "X", 3, 0, true, true,
	},
	{
		63, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-14 22:00:00", "X", 4, 0, false, true,
	},
	{
		64, "待定", "待定",
		homeTeamWinOdds, awayTeamWinOdds, tiedOdds,
		"2018-07-15 23:00:00", "X", 5, 0, true, true,
	},
}

const (
	reqURL = "http://localhost:9614/update_schedule"
)

func updateAll() {
	for _, schedule := range schedules {
		jsonData, err := json.Marshal(schedule)
		if err != nil {
			log.Fatalf("json marshal error: %v\n", err)
		}
		req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonData))
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

func updateSchedule(id int, status int, enableDisplay bool, disableDetting bool) {
	schedule := schedules[id-1]
	schedule.ScheduleStatus = status
	schedule.EnableDisplay = enableDisplay
	schedule.DisableBetting = disableDetting
	jsonData, err := json.Marshal(schedule)
	if err != nil {
		log.Fatalf("json marshal error: %v\n", err)
	}
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("do post request failed, error: %v, body: %v\n", err, body)
	}
}

func main() {
	updateSchedule(1, 1, false, false)
	//updateAll()
}
