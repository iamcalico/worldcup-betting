package main

type ScheduleType int

const (
	GroupMatches       ScheduleType = 0 // 小组赛
	RoundEight                          // 八强赛
	FinalFour                           // 四强赛
	Semifinal                           // 半决赛
	MatchForThirdPlace                  // 季军赛
	Finals                              // 总决赛
	All
)

type ScheduleStatus int

const (
	NotStarted  ScheduleStatus = 0 // 未开始
	HomeTeamWin                    // 主队胜利
	AwayTeamWin                    // 客队胜利
	Draw                           // 平局
)

type BetStatus int

const (
	BetNotFinish = 0
	WinBet       = 1
	LostBet      = 2
)

type CountryInfo struct {
	CountryID     int    `json:"id"`
	CountryName   string `json:"name"`
	CountryPicURL string `json:"logo"`
}

type AddTipsRequest struct {
	TipsList []Tips `json:"tips"`
}

// 每个 Schedule 代表一场世界杯赛事
type Schedule struct {
	ScheduleID      int            `json:"schedule_id"`        // 赛事 ID，用以唯一标识每场比赛
	HomeTeam        string         `json:"home_team"`          // 主队
	AwayTeam        string         `json:"away_team"`          // 客队
	HomeTeamWinOdds float64        `json:"home_team_win_odds"` // 主队胜利的赔率
	AwayTeamWinOdds float64        `json:"away_team_win_odds"` // 客队胜利的赔率
	TiedOdds        float64        `json:"tied_odds"`          // 平局的赔率
	ScheduleTime    string         `json:"schedule_time"`      // 比赛时间
	ScheduleGroup   string         `json:"schedule_group"`     // 比赛组别
	ScheduleType    ScheduleType   `json:"schedule_type"`      // 比赛类别
	ScheduleStatus  ScheduleStatus `json:"schedule_status"`    // 比赛状态
	DisableBetting  bool           `json:"disable_betting"`    // 是否允许投注
	EnableDisplay   bool           `json:"enable_dispaly"`     // 是否显示在投注页
}

type Schedule2 struct {
	ScheduleID      int            `json:"schedule_id"`               // 赛事 ID，用以唯一标识每场比赛
	HomeTeam        int            `json:"home_team"`                 // 主队
	AwayTeam        int            `json:"away_team"`                 // 客队
	HomeTeamWinOdds float64        `json:"home_team_win_odds"`        // 主队胜利的赔率
	AwayTeamWinOdds float64        `json:"away_team_win_odds"`        // 客队胜利的赔率
	TiedOdds        float64        `json:"tied_odds"`                 // 平局的赔率
	ScheduleTime    string         `json:"schedule_time"`             // 比赛时间
	ScheduleGroup   string         `json:"schedule_group"`            // 比赛组别
	ScheduleType    ScheduleType   `json:"schedule_type"`             // 比赛类别
	ScheduleStatus  ScheduleStatus `json:"schedule_status,omitempty"` // 比赛状态
	DisableBetting  bool           `json:"disable_betting"`           // 是否允许投注
	EnableDisplay   bool           `json:"enable_dispaly"`            // 是否显示在投注页
}

type User struct {
	UserId              int     `json:"user_id"`
	EnglishName         string  `json:"en_name"`
	ChineseName         string  `json:"cn_name"`
	Password            string  `json:"password"`
	Money               float64 `json:"money"`
	EnableResetPassword bool    `json:"enable_reset_password"`
	LastLoginTime       string  `json:"last_login_time"`
	WinCount            int     `json:"omitempty"`
	BetCount            int     `json:"bet_count"`
}

type RewardHistory struct {
	UserId      int    `json:"user_id"`
	RewardTime  string `json:"reward_time"`
	RewardMoney int    `json:"reward_money"`
}

type RankRsp struct {
	UserID      int     `json:"user_id"`
	RTXName     string  `json:"en_name"`
	ChineseName string  `json:"cn_name"`
	Money       float64 `json:"money"`
	WinCount    int     `json:"win_count"`
	BetCount    int     `json:"bet_count"`
}

type DailyRewardRequest struct {
	UserID int `json:"user_id"`
}

type BetRequest struct {
	UserId        int     `json:"user_id"`
	ScheduleId    int     `json:"schedule_id"`
	BettingMoney  int     `json:"betting_money"`
	BettingResult int     `json:"betting_result"`
	BettingOdds   float64 `json:"betting_odds"`
	BettingStatus int     `json:"bet_status"`
	WinMoney      float64 `json:"win_money"`
}

type AuthorizeRequest struct {
	ChineseName string `json:"ch_name"`  // 中文名
	EnglishName string `json:"en_name"`  // 英文名
	Password    string `json:"password"` // MD5 之后的密码
}

type Config struct {
	MySQLUser        string `mapstructure:"mysql_server"`
	MySQLPassword    string `mapstructure:"mysql_password"`
	MySQLNet         string `mapstructure:"mysql_net"`
	MySQLDBName      string `mapstructure:"mysql_db_name"`
	MySQLAddr        string `mapstructure:"mysql_addr"`
	CSVNameList      string `mapstructure:"csv_name_list"`
	InitialMoney     int    `mapstructure:"initial_money"`
	DailyRewardMoney int    `mapstructure:"daily_reward_money"`
	EnableWhiteList  bool   `mapstructure:"enable_white_list"`
	DomainName       string `mapstructure:"domain_name"`
	TimiNewUser      string `mapstructure:"timi_new_user"`
	ServerPort       string `mapstructure:"server_port"`
}

type Tips struct {
	TipsID        int    `json:"tips_id"`
	Content       string `json:"content"`
	EnableDisplay bool   `json:"enable_display"`
}

type AddNewUserReq struct {
	ChineseName string `json:"ch_name"`
	EnglishName string `json:"en_name"`
}

type GrantResetPassword struct {
	ChineseName string `json:"ch_name"`
	EnglishName string `json:"en_name"`
}
