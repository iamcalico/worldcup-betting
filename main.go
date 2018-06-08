package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

var (
	db *sql.DB
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

// TODO: 应该将这个参数弄成配置文件
const (
	timiUserCSVFile = "timi_users.csv"
	defaultMoney    = 5000
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

type User struct {
	UserId              int     `json:"user_id"`
	RTXName             string  `json:"rtx_name"`
	ChineseName         string  `json:"cn_name"`
	Password            string  `json:"password"`
	Money               float64 `json:"money"`
	EnableResetPassword bool
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
		Addr:                 "10.211.55.18",
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

func schedules(db *sql.DB, scheduleType ScheduleType) ([]Schedule, error) {
	if db == nil {
		return []Schedule{}, fmt.Errorf("db handle is nil")
	}

	var schedules []Schedule
	rows, err := db.Query("SELECT * FROM schedule WHERE schedule_type = ?", scheduleType)
	if err != nil {
		log.Fatalf("query db failed, error: %v\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		var schedule Schedule
		err := rows.Scan(&schedule.ScheduleID, &schedule.HomeTeam, &schedule.AwayTeam,
			&schedule.HomeTeamWinOdds, &schedule.AwayTeamWinOdds, &schedule.TiedOdds,
			&schedule.ScheduleTime, &schedule.ScheduleGroup, &schedule.ScheduleType,
			&schedule.ScheduleStatus, &schedule.DisableBetting)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan rows: %v failed, error: %v\n", rows, err)
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

func handleSchedules(c *gin.Context) {
	scheduleType := c.Query("type")

	queryScheduleType, err := strconv.Atoi(scheduleType)
	if err != nil {
		illegalParametersRsp(c)
		return
	}

	queryScheduleType2 := ScheduleType(queryScheduleType)
	if queryScheduleType2 < GroupMatches || queryScheduleType2 > Finals {
		illegalParametersRsp(c)
		return
	}

	schedules, err := schedules(db, queryScheduleType2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get schedules failed, error: %v\n", err)
		queryMySQLFailedRsp(c)
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

	rows, err := db.Query("SELECT * FROM schedule WHERE schedule_id = ?", schedule.ScheduleID)
	if err != nil {
		queryMySQLFailedRsp(c)
		fmt.Fprintf(os.Stderr, "query schedule failed, err: %v\n", err)
		return
	}

	// TODO: 遍历这场比赛下注的人，更新其赔率的情况
	if rows.Next() {
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
			updateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "update schedule failed, result:%v, err: %v\n", result, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK"})
		return
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
		insertMySQLFailedRsp(c)
		fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
		return
	}
	result, err := stmt.Exec(schedule.HomeTeam, schedule.AwayTeam,
		schedule.HomeTeamWinOdds, schedule.AwayTeamWinOdds, schedule.TiedOdds,
		schedule.ScheduleTime, schedule.ScheduleGroup, schedule.ScheduleType,
		schedule.ScheduleStatus, schedule.DisableBetting)
	if err != nil {
		insertMySQLFailedRsp(c)
		fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
		return
	}
	lastId, err := result.LastInsertId()
	if err != nil {
		insertMySQLFailedRsp(c)
		fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK", "schedule_id": lastId})
}

// TODO: 必须先验证下用户下注的合法性，比如有没有那么多金币
func handleBet(c *gin.Context) {
	var betRequest BetRequest
	if c.Bind(&betRequest) != nil {
		illegalParametersRsp(c)
		return
	}

	// 验证用户是否已经对这场比赛下过注
	rows, err := db.Query("SELECT * FROM bet WHERE user_id = ? and schedule_id = ?",
		betRequest.UserId, betRequest.ScheduleId)
	if rows.Next() {
		alreadyBet(c)
		return
	}

	// TODO: 验证这场赛事已经可以下注

	// 验证用户是否有足够的钱进行下注
	rows, err = db.Query("SELECT * FROM user WHERE user_id = ?", betRequest.UserId)
	if err != nil {
		queryMySQLFailedRsp(c)
		fmt.Fprintf(os.Stderr, "query mysql failed, err: %v\n", err)
		return
	}
	if rows.Next() {
		var user User
		err := rows.Scan(&user.UserId, &user.RTXName, &user.ChineseName, &user.Password, &user.Money)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan rows: %v failed, error: %v\n", rows, err)
			queryUserFailedRsp(c)
			return
		}

		// 如果用户已经没有足够的钱下注
		if int(user.Money) < betRequest.BettingMoney {
			notEnoughMoney(c)
			return
		}

		// 在 bet 表插入相应的记录
		stmt, err := db.Prepare("INSERT INTO " +
			"bet(user_id,schedule_id,betting_money,betting_result,betting_odds) " +
			"VALUES (?,?,?,?,?)")
		if err != nil {
			insertMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
			return
		}
		result, err := stmt.Exec(betRequest.UserId, betRequest.ScheduleId,
			betRequest.BettingMoney, betRequest.BettingResult, betRequest.BettingOdds)
		if err != nil {
			insertMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
			return
		}

		// 更新 user 表中用户的金币（正确来说，应该这几个操作是一个 事务 操作）
		stmt, err = db.Prepare("UPDATE user SET money = ? WHERE user_id = ?")
		if err != nil {
			updateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
			return
		}
		result, err = stmt.Exec(int(user.Money)-betRequest.BettingMoney, betRequest.UserId)
		if err != nil {
			updateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "update schedule failed, result:%v, err: %v\n", result, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK"})
	} else {
		userNotExist(c)
		return
	}
}

func handleAuthorize(c *gin.Context) {
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
		err := rows.Scan(&user.UserId, &user.RTXName, &user.ChineseName, &user.Password, &user.Money)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan rows: %v failed, error: %v\n", rows, err)
			queryUserFailedRsp(c)
			return
		}
		if user.Password != authorizeRequest.Password {
			fmt.Fprintf(os.Stderr, "incorrect password, [ch_name:%v, en_name:%v], want: %v, got: %v\n",
				user.ChineseName, user.RTXName, user.Password, authorizeRequest.Password)
			incorrectPasswordRsp(c)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK", "user_id": user.UserId, "money": user.Money})
	} else {
		// 说明是第一次登陆
		stmt, err := db.Prepare("INSERT INTO " +
			"user(rtx_name,chinese_name,password,money) " +
			"VALUES (?,?,?,?)")
		if err != nil {
			insertMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
			return
		}
		result, err := stmt.Exec(authorizeRequest.EnglishName, authorizeRequest.ChineseName, authorizeRequest.Password, defaultMoney)
		if err != nil {
			insertMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
			return
		}
		lastId, err := result.LastInsertId()
		if err != nil {
			insertMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK", "user_id": lastId, "money": defaultMoney})
	}
}

func handleBettingHistory(c *gin.Context) {
	userIDPara := c.Query("user_id")

	userID, err := strconv.Atoi(userIDPara)
	if err != nil {
		illegalParametersRsp(c)
		return
	}

	var betHistory []BetRequest
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
		err := rows.Scan(&user.UserId, &user.RTXName, &user.ChineseName, &user.Password, &user.Money, &user.EnableResetPassword)
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

func handleGrantResetPassword(c *gin.Context) {

}

// TODO: 还有一个每天送金币的功能
func init() {
	db = sqlDB()
	readUserFile(timiUserCSVFile)
}

// TODO: 加一个授权更新密码的接口
func main() {
	router := gin.Default()
	router.GET("/schedules", handleSchedules)
	router.GET("/betting_history", handleBettingHistory)
	router.PUT("/new_schedule", handleNewSchedule)
	router.POST("/update_schedule", handleUpdateSchedule)
	router.POST("/bet", handleBet)
	router.POST("/authorize", handleAuthorize)
	router.POST("/reset_password", handleResetPassword)
	router.POST("/grant_reset_password", handleGrantResetPassword)
	router.Run(":9614")
}
