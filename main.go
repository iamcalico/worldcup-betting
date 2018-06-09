package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

var (
	db         *sql.DB
	countryMap = map[string]int{
		"俄罗斯":   1,
		"沙特阿拉伯": 2,
		"埃及":    3,
		"乌拉圭":   4,
		"摩洛哥":   5,
		"伊朗":    6,
		"葡萄牙":   7,
		"西班牙":   8,
		"法国":    9,
		"澳大利亚":  10,
		"阿根廷":   11,
		"冰岛":    12,
		"秘鲁":    13,
		"丹麦":    14,
		"克罗地亚":  15,
		"尼日利亚":  16,
		"哥斯达黎加": 17,
		"塞尔维亚":  18,
		"德国":    19,
		"墨西哥":   20,
		"巴西":    21,
		"瑞士":    22,
		"瑞典":    23,
		"韩国":    24,
		"比利时":   25,
		"巴拿马":   26,
		"突尼斯":   27,
		"英格兰":   28,
		"哥伦比亚":  29,
		"日本":    30,
		"波兰":    31,
		"塞纳加尔":  32,
	}
)

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

// TODO: 应该将这个参数弄成配置文件
const (
	timiUserCSVFile         = "timi_users.csv"
	defaultMoney            = 5000
	defaultRewardDailyMoney = 50
)

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
}

type User struct {
	UserId              int     `json:"user_id"`
	EnglishName         string  `json:"en_name"`
	ChineseName         string  `json:"cn_name"`
	Password            string  `json:"password"`
	Money               float64 `json:"money"`
	EnableResetPassword bool    `json:"enable_reset_password"`
	LastLoginTime       string  `json:"last_login_time"`
}

type RewardHistory struct {
	UserId      int    `json:"user_id"`
	RewardTime  string `json:"reward_time"`
	RewardMoney int    `json:"reward_money"`
}

type BetRequest struct {
	UserId        int     `json:"user_id"`
	ScheduleId    int     `json:"schedule_id"`
	BettingMoney  int     `json:"betting_money"`
	BettingResult int     `json:"betting_result"`
	BettingOdds   float64 `json:"betting_odds"`
}

type AuthorizeRequest struct {
	ChineseName string `json:"ch_name"`  // 中文名
	EnglishName string `json:"en_name"`  // 英文名
	Password    string `json:"password"` // MD5 之后的密码
}

// TODO: 配置信息会以配置文件的形式呈现
func sqlDB() *sql.DB {
	cfg := mysql.Config{
		User:                 "root",
		Passwd:               "123456",
		Net:                  "tcp",
		Addr:                 "localhost",
		DBName:               "betting",
		AllowNativePasswords: true,
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatalf("open sql db failed, error: %v\n", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("db ping test failed, error: %v\n", err)
	}
	return db
}

func schedules(db *sql.DB, scheduleType ScheduleType) ([]Schedule2, error) {
	var (
		schedules []Schedule2
		rows      *sql.Rows
		err       error
	)

	if scheduleType == All {
		rows, err = db.Query("SELECT * FROM schedule")
	} else {
		rows, err = db.Query("SELECT * FROM schedule WHERE schedule_type = ?", scheduleType)
	}
	if err != nil {
		return []Schedule2{}, err
	}

	for rows.Next() {
		var (
			schedule  Schedule
			schedule2 Schedule2
		)
		err := rows.Scan(&schedule.ScheduleID, &schedule.HomeTeam, &schedule.AwayTeam,
			&schedule.HomeTeamWinOdds, &schedule.AwayTeamWinOdds, &schedule.TiedOdds,
			&schedule.ScheduleTime, &schedule.ScheduleGroup, &schedule.ScheduleType,
			&schedule.ScheduleStatus, &schedule.DisableBetting)
		if err != nil {
			return []Schedule2{}, err
		}

		schedule2.ScheduleID = schedule.ScheduleID
		schedule2.HomeTeam = countryToID(schedule.HomeTeam)
		schedule2.AwayTeam = countryToID(schedule.AwayTeam)
		schedule2.HomeTeamWinOdds = schedule.HomeTeamWinOdds
		schedule2.AwayTeamWinOdds = schedule.AwayTeamWinOdds
		schedule2.TiedOdds = schedule.TiedOdds
		schedule2.ScheduleTime = schedule.ScheduleTime
		schedule2.ScheduleGroup = schedule.ScheduleGroup
		schedule2.ScheduleType = schedule.ScheduleType
		schedule2.DisableBetting = schedule.DisableBetting
		schedule2.ScheduleStatus = schedule.ScheduleStatus

		schedules = append(schedules, schedule2)
	}

	return schedules, nil
}

func countryToID(country string) int {
	return countryMap[country]
}

func handleSchedules(c *gin.Context) {
	scheduleType := c.Query("type")

	var queryScheduleType ScheduleType
	if scheduleType == "" {
		queryScheduleType = All
	} else {
		scheduleType2, err := strconv.Atoi(scheduleType)
		if err != nil {
			illegalParametersRsp(c)
			return
		}
		queryScheduleType = ScheduleType(scheduleType2)
		if queryScheduleType < GroupMatches || queryScheduleType > All {
			illegalParametersRsp(c)
			return
		}
	}

	schedules, err := schedules(db, queryScheduleType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get schedules failed, error: %v\n", err)
		operateMySQLFailedRsp(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    0,
		"desc":      "OK",
		"schedules": schedules,
	})
}

// TODO: 这种回复一条 JSON 后就退出的操作有没有什么更优雅的做法
func handleUpdateSchedule(c *gin.Context) {
	var schedule Schedule
	if c.Bind(&schedule) != nil {
		illegalParametersRsp(c)
		return
	}

	// 先验证要更新的赛程是在数据库中
	rows, err := db.Query("SELECT * FROM schedule WHERE schedule_id = ?", schedule.ScheduleID)
	if err != nil {
		operateMySQLFailedRsp(c)
		fmt.Fprintf(os.Stderr, "query schedule failed, err: %v\n", err)
		return
	}
	if rows.Next() {
		// 验证赛事的结果的合法性
		if ScheduleStatus(schedule.ScheduleStatus) >= HomeTeamWin || ScheduleStatus(schedule.ScheduleStatus) <= Draw {
			stmt, err := db.Prepare("UPDATE schedule SET home_team = ?, away_team = ?, " +
				"home_team_win_odds = ?, away_team_win_odds = ?, tied_odds = ?, " +
				"schedule_time = ?, schedule_group = ?, schedule_type = ?, " +
				"schedule_status = ?, disable_betting = ? " + "WHERE schedule_id = ?")
			if err != nil {
				updateMySQLFailedRsp(c)
				fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
				return
			}
			result, err := stmt.Exec(schedule.HomeTeam, schedule.AwayTeam,
				schedule.HomeTeamWinOdds, schedule.AwayTeamWinOdds, schedule.TiedOdds,
				schedule.ScheduleTime, schedule.ScheduleGroup, schedule.ScheduleType, schedule.ScheduleStatus,
				schedule.DisableBetting, schedule.ScheduleID)
			if err != nil {
				operateMySQLFailedRsp(c)
				fmt.Fprintf(os.Stderr, "update schedule failed, result:%v, err: %v\n", result, err)
				return
			}

			// 遍历竞猜表，计算出每个人竞猜的结果，如果竞猜成功，增加相应用户的金币数
			var betRequest BetRequest
			rows, _ := db.Query("SELECT * FROM bet WHERE schedule_id = ?", schedule.ScheduleID)
			for rows.Next() {
				rows.Scan(&betRequest.UserId, &betRequest.ScheduleId, &betRequest.BettingMoney, &betRequest.BettingResult, &betRequest.BettingOdds)
				if ScheduleStatus(betRequest.BettingResult) == ScheduleStatus(schedule.ScheduleStatus) {
					var money float64
					rows, _ := db.Query("SELECT money FROM user WHERE user_id = ?", betRequest.UserId)
					if rows.Next() {
						rows.Scan(&money)
					}
					stmt, _ := db.Prepare("UPDATE user SET money = ? WHERE user_id = ?")
					currentMoney := float64(betRequest.BettingMoney)*betRequest.BettingOdds + money
					stmt.Exec(currentMoney, betRequest.UserId)
				}
			}

			c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK"})
			return
		} else {
			illegalParametersRsp(c)
			return
		}

	} else { // 赛事必须已在 schedule 表中才能更新成功
		scheduleNotExistRsp(c)
		return
	}
}

func handleNewSchedule(c *gin.Context) {
	var schedule Schedule
	if c.Bind(&schedule) != nil {
		illegalParametersRsp(c)
		return
	}
	stmt, err := db.Prepare("INSERT INTO " +
		"schedule(home_team,away_team,home_team_win_odds,away_team_win_odds,tied_odds,schedule_time,schedule_group,schedule_type,schedule_status,disable_betting) " +
		"VALUES (?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		operateMySQLFailedRsp(c)
		fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
		return
	}
	result, err := stmt.Exec(schedule.HomeTeam, schedule.AwayTeam,
		schedule.HomeTeamWinOdds, schedule.AwayTeamWinOdds, schedule.TiedOdds,
		schedule.ScheduleTime, schedule.ScheduleGroup, schedule.ScheduleType,
		schedule.ScheduleStatus, schedule.DisableBetting)
	if err != nil {
		operateMySQLFailedRsp(c)
		fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
		return
	}
	lastId, err := result.LastInsertId()
	if err != nil {
		operateMySQLFailedRsp(c)
		fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK", "schedule_id": lastId})
}

func handleBet(c *gin.Context) {
	var betRequest BetRequest
	if c.Bind(&betRequest) != nil {
		illegalParametersRsp(c)
		return
	}

	// 验证这场赛事已经可以下注
	var disableBetting bool
	rows, err := db.Query("SELECT disable_betting FROM schedule WHERE schedule_id = ?", betRequest.ScheduleId)
	if err != nil {
		operateMySQLFailedRsp(c)
		fmt.Fprintf(os.Stderr, "query mysql failed, err: %v\n", err)
		return
	}
	if rows.Next() {
		err := rows.Scan(&disableBetting)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan rows: %v failed, error: %v\n", rows, err)
			operateMySQLFailedRsp(c)
			return
		}
		if disableBetting {
			disableBet(c)
			return
		} else {
			// 验证用户是否已经对这场比赛下过注
			rows, err := db.Query("SELECT * FROM bet WHERE user_id = ? and schedule_id = ?",
				betRequest.UserId, betRequest.ScheduleId)
			if rows.Next() {
				alreadyBet(c)
				return
			}

			// 验证用户是否有足够的钱进行下注
			rows, err = db.Query("SELECT money FROM user WHERE user_id = ?", betRequest.UserId)
			if err != nil {
				operateMySQLFailedRsp(c)
				fmt.Fprintf(os.Stderr, "query mysql failed, err: %v\n", err)
				return
			}
			if rows.Next() {
				var money float64
				err := rows.Scan(&money)
				if err != nil {
					fmt.Fprintf(os.Stderr, "scan rows: %v failed, error: %v\n", rows, err)
					operateMySQLFailedRsp(c)
					return
				}

				// 如果用户已经没有足够的钱下注
				if int(money) < betRequest.BettingMoney {
					notEnoughMoney(c)
					return
				}

				// 在 bet 表插入相应的记录
				stmt, err := db.Prepare("INSERT INTO " +
					"bet(user_id,schedule_id,betting_money,betting_result,betting_odds) " +
					"VALUES (?,?,?,?,?)")
				if err != nil {
					operateMySQLFailedRsp(c)
					fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
					return
				}
				result, err := stmt.Exec(betRequest.UserId, betRequest.ScheduleId,
					betRequest.BettingMoney, betRequest.BettingResult, betRequest.BettingOdds)
				if err != nil {
					operateMySQLFailedRsp(c)
					fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
					return
				}

				// 更新 user 表中用户的金币（正确来说，应该这几个操作是一个 事务 操作）
				stmt, err = db.Prepare("UPDATE user SET money = ? WHERE user_id = ?")
				if err != nil {
					operateMySQLFailedRsp(c)
					fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
					return
				}
				result, err = stmt.Exec(int(money)-betRequest.BettingMoney, betRequest.UserId)
				if err != nil {
					operateMySQLFailedRsp(c)
					fmt.Fprintf(os.Stderr, "update schedule failed, result:%v, err: %v\n", result, err)
					return
				}

				c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK"})
			} else {
				userNotExist(c)
				return
			}
		}
	} else {
		scheduleNotExistRsp(c)
		return
	}
}

func handleAuthorize(c *gin.Context) {
	var authorizeRequest AuthorizeRequest
	if c.Bind(&authorizeRequest) != nil {
		illegalParametersRsp(c)
		return
	}

	// 必须验证用户在白名单之内
	if isIllegalUser(authorizeRequest.ChineseName, authorizeRequest.EnglishName) != true {
		illegalUserRsp(c)
		return
	}

	// 判读是否第一次登陆，如果是第一次登陆，则数据库中找不到相应的记录
	rows, err := db.Query("SELECT * FROM user WHERE chinese_name = ? and rtx_name = ?",
		authorizeRequest.ChineseName, authorizeRequest.EnglishName)
	if err != nil {
		queryMySQLFailedRsp(c)
		fmt.Fprintf(os.Stderr, "query schedule failed, err: %v\n", err)
		return
	}

	loginTimeStamp := time.Now().Unix()
	tm := time.Unix(loginTimeStamp, 0)
	loginTime := tm.Format("2006-01-02 15:04:05")
	if rows.Next() {
		var user User
		err := rows.Scan(&user.UserId, &user.EnglishName, &user.ChineseName, &user.Password, &user.Money, &user.EnableResetPassword, &user.LastLoginTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan rows: %v failed, error: %v\n", rows, err)
			queryUserFailedRsp(c)
			return
		}
		if user.Password != authorizeRequest.Password {
			fmt.Fprintf(os.Stderr, "incorrect password, [ch_name:%v, en_name:%v], want: %v, got: %v\n",
				user.ChineseName, user.EnglishName, user.Password, authorizeRequest.Password)
			incorrectPasswordRsp(c)
			return
		}

		isReward, rewardMoney := dailyReward(user.UserId, loginTimeStamp)
		if isReward {
			user.Money += float64(rewardMoney)
		}
		// 更新登陆时间
		stmt, err := db.Prepare("UPDATE user SET money = ?, last_login_time = ? WHERE user_id = ?")
		if err != nil {
			updateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
			return
		}
		result, err := stmt.Exec(user.Money, loginTime, user.UserId)
		if err != nil {
			updateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "update schedule failed, result:%v, err: %v\n", result, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK", "user_id": user.UserId, "money": user.Money, "daily_reward": rewardMoney})
	} else {
		// 说明是第一次登陆
		stmt, err := db.Prepare("INSERT INTO " +
			"user(rtx_name,chinese_name,password,money,enable_reset_password,last_login_time) " +
			"VALUES (?,?,?,?,?,?)")
		if err != nil {
			operateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
			return
		}
		result, err := stmt.Exec(authorizeRequest.EnglishName, authorizeRequest.ChineseName, authorizeRequest.Password,
			defaultMoney, false, loginTime)
		if err != nil {
			operateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
			return
		}
		lastId, err := result.LastInsertId()
		if err != nil {
			operateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK", "user_id": lastId, "money": defaultMoney, "first_login": true})
	}
}

func dailyReward(userID int, loginTimeStamp int64) (bool, int) {
	tm := time.Unix(loginTimeStamp, 0)
	lowerBound := tm.Format("2006-01-02") + " 00:00:00"
	upperBound := tm.Format("2006-01-02") + " 23:59:59"
	rows, err := db.Query("SELECT * FROM reward WHERE user_id = ? and reward_time > ? and reward_time < ?",
		userID, lowerBound, upperBound)
	if err != nil {
		log.Fatalf("query db failed, error: %v\n", err)
	}

	// 如果查到 reward 表中已经有了记录，说明今天已经送过金币
	if rows.Next() {
		return false, 0
	} else {
		// 说明今天还没有送金币
		stmt, _ := db.Prepare("INSERT INTO " + "reward(user_id, reward_time, reward_money) " + "VALUES (?,?,?)")
		_, err = stmt.Exec(userID, tm.Format("2006-01-02 15:03:04"), defaultRewardDailyMoney)
		if err != nil {
			return false, 0
		}
		return true, defaultRewardDailyMoney
	}
}

func handleBettingHistory(c *gin.Context) {
	userIDPara := c.Query("user_id")

	userID, err := strconv.Atoi(userIDPara)
	if err != nil {
		illegalParametersRsp(c)
		return
	}

	betHistory := []BetRequest{}
	rows, err := db.Query("SELECT * FROM bet WHERE user_id = ?", userID)
	if err != nil {
		log.Fatalf("query db failed, error: %v\n", err)
	}
	for rows.Next() {
		var betRequest BetRequest
		err := rows.Scan(&betRequest.UserId, &betRequest.ScheduleId,
			&betRequest.BettingMoney, &betRequest.BettingResult, &betRequest.BettingOdds)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan rows: %v failed, error: %v\n", rows, err)
		}
		betHistory = append(betHistory, betRequest)
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK", "betting_history": betHistory})
}

func handleResetPassword(c *gin.Context) {
	var authorizeRequest AuthorizeRequest
	if c.Bind(&authorizeRequest) != nil {
		illegalParametersRsp(c)
		return
	}

	if isIllegalUser(authorizeRequest.ChineseName, authorizeRequest.EnglishName) != true {
		illegalUserRsp(c)
		return
	}
	rows, err := db.Query("SELECT * FROM user WHERE chinese_name = ? and rtx_name = ?",
		authorizeRequest.ChineseName, authorizeRequest.EnglishName)
	if err != nil {
		queryMySQLFailedRsp(c)
		fmt.Fprintf(os.Stderr, "query schedule failed, err: %v\n", err)
		return
	}

	if rows.Next() {
		var user User
		err := rows.Scan(&user.UserId, &user.EnglishName, &user.ChineseName, &user.Password, &user.Money, &user.EnableResetPassword)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan rows: %v failed, error: %v\n", rows, err)
			queryUserFailedRsp(c)
			return
		}
		if user.EnableResetPassword == true {
			// 重置密码并登陆
			stmt, err := db.Prepare("UPDATE user SET password = ?, enable_reset_password = ? WHERE user_id = ?")
			if err != nil {
				updateMySQLFailedRsp(c)
				fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
				return
			}
			result, err := stmt.Exec(authorizeRequest.Password, false, user.UserId)
			if err != nil {
				updateMySQLFailedRsp(c)
				fmt.Fprintf(os.Stderr, "update schedule failed, result:%v, err: %v\n", result, err)
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK", "user_id": user.UserId, "money": user.Money})
		} else {
			notAllowResetPassword(c)
			return
		}
	} else {
		userNotExist(c)
		return
	}
}

// TODO: 加一个授权更新密码的接口
func handleGrantResetPassword(c *gin.Context) {

}

type RankRsp struct {
	RTXName     string  `json:"rtx_name"`
	ChineseName string  `json:"cn_name"`
	Money       float64 `json:"money"`
}

func handleRank(c *gin.Context) {
	limit := c.Query("limit")
	if limit == "" {
		limit = "20"
	}

	ranks := []RankRsp{}
	rows, err := db.Query("SELECT rtx_name, chinese_name, money FROM user ORDER BY money limit ?", limit)
	if err != nil {
		queryMySQLFailedRsp(c)
		return
	}

	for rows.Next() {
		var rank RankRsp
		err := rows.Scan(&rank.RTXName, &rank.ChineseName, &rank.Money)
		if err != nil {
			queryMySQLFailedRsp(c)
			return
		}
		ranks = append(ranks, rank)
	}
	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"desc":   "OK",
		"rank":   ranks,
	})
}

func handleRewardHistory(c *gin.Context) {
	userID := c.Query("user_id")

	rewardHistory := []RewardHistory{}
	rows, _ := db.Query("SELECT * FROM reward WHERE user_id = ?", userID)
	for rows.Next() {
		var history RewardHistory
		err := rows.Scan(&history.UserId, &history.RewardTime, &history.RewardMoney)
		if err != nil {
			queryMySQLFailedRsp(c)
			return
		}
		rewardHistory = append(rewardHistory, history)
	}
	c.JSON(http.StatusOK, gin.H{
		"status":         0,
		"desc":           "OK",
		"reward_history": rewardHistory,
	})
}

func handleMyInfo(c *gin.Context) {
	userID := c.Query("user_id")
	rows, _ := db.Query("SELECT user_id,rtx_name,chinese_name,money FROM user WHERE user_id = ?", userID)
	if rows.Next() {
		var user User
		err := rows.Scan(&user.UserId, &user.EnglishName, &user.ChineseName, &user.Money)
		if err != nil {
			operateMySQLFailedRsp(c)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  0,
			"desc":    "OK",
			"id":      user.UserId,
			"money":   user.Money,
			"cn_name": user.ChineseName,
			"en_name": user.EnglishName,
		})
	} else {
		userNotExist(c)
		return
	}
}

func init() {
	db = sqlDB()
	readUserFile(timiUserCSVFile)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, XMLHttpRequest, "+
			"Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.String(200, "ok")
			return
		}
		c.Next()
	}
}

func main() {
	router := gin.Default()
	router.Use(CORSMiddleware())

	router.PUT("/new_schedule", handleNewSchedule)

	router.GET("/schedules", handleSchedules)
	router.GET("/rank", handleRank)
	router.GET("/betting_history", handleBettingHistory)
	router.GET("/reward_history", handleRewardHistory)
	router.GET("/my", handleMyInfo)

	router.POST("/update_schedule", handleUpdateSchedule)
	router.POST("/bet", handleBet)
	router.POST("/authorize", handleAuthorize)
	router.POST("/reset_password", handleResetPassword)
	router.POST("/grant_reset_password", handleGrantResetPassword)
	router.Run(":9614")
}
