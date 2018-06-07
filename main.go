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

type BetRequest struct {
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
	schedules := make([]Schedule, 1)
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
			log.Fatalf("scan rows: %v failed, error: %v\n", rows, err)
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

func illegalParametersRsp(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status": 1,
		"desc":   "Illegal Parameters",
	})
}

func scheduleNotExistRsp(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status": 2,
		"desc":   "Schedule is not exist",
	})
}

func queryScheduleFailedRsp(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"status": 3,
		"desc":   "Query schedule failed",
	})
}

func updateScheduleFailedRsp(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"status": 4,
		"desc":   "Update schedule failed",
	})
}

func insertScheduleFailedRsp(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"status": 5,
		"desc":   "Insert schedule failed",
	})
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
		log.Fatalf("%v\n", err)
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
		queryScheduleFailedRsp(c)
		fmt.Fprintf(os.Stderr, "query schedule failed, err: %v\n", err)
		return
	}
	if rows.Next() {
		stmt, err := db.Prepare("UPDATE schedule SET home_team = ?, away_team = ?, " +
			"home_team_win_odds = ?, away_team_win_odds = ?, tied_odds = ?, " +
			"schedule_time = ?, schedule_group = ?, schedule_type = ?, " +
			"schedule_status = ?, disable_betting = ? " + "WHERE schedule_id = ?")
		if err != nil {
			updateScheduleFailedRsp(c)
			fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
			return
		}
		result, err := stmt.Exec(schedule.HomeTeam, schedule.AwayTeam,
			schedule.HomeTeamWinOdds, schedule.AwayTeamWinOdds, schedule.TiedOdds,
			schedule.ScheduleTime, schedule.ScheduleGroup, schedule.ScheduleType, schedule.ScheduleStatus,
			schedule.DisableBetting, schedule.ScheduleID)
		if err != nil {
			updateScheduleFailedRsp(c)
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
		insertScheduleFailedRsp(c)
		fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
		return
	}
	result, err := stmt.Exec(schedule.HomeTeam, schedule.AwayTeam,
		schedule.HomeTeamWinOdds, schedule.AwayTeamWinOdds, schedule.TiedOdds,
		schedule.ScheduleTime, schedule.ScheduleGroup, schedule.ScheduleType,
		schedule.ScheduleStatus, schedule.DisableBetting)
	if err != nil {
		insertScheduleFailedRsp(c)
		fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
		return
	}
	lastId, err := result.LastInsertId()
	if err != nil {
		insertScheduleFailedRsp(c)
		fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK", "schedule_id": lastId})
}

func handleBet(c *gin.Context) {

}

func handleAuthorize(c *gin.Context) {

}

func handleBettingHistory(c *gin.Context) {

}

func init() {
	db = sqlDB()
}

func main() {
	router := gin.Default()
	router.GET("/schedules", handleSchedules)
	router.GET("/betting_history", handleBettingHistory)
	router.PUT("/new_schedule", handleNewSchedule)
	router.POST("/update_schedule", handleUpdateSchedule)
	router.POST("/bet", handleBet)
	router.POST("/authorize", handleAuthorize)
	router.Run(":9614")
}
