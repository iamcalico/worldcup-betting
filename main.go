package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

func sqlDB() *sql.DB {
	cfg := mysql.Config{
		User:                 config.MySQLUser,
		Passwd:               config.MySQLPassword,
		Net:                  config.MySQLNet,
		Addr:                 config.MySQLAddr,
		DBName:               config.MySQLDBName,
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

func parseConfig() {
	configPath := flag.String("config", "./config.toml", "Specify the config file")
	flag.Parse()
	v := viper.New()
	basename := filepath.Base(*configPath)
	v.SetConfigName(basename[0:(len(basename) - len(filepath.Ext(basename)))])
	v.AddConfigPath(filepath.Dir(*configPath))
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("read config error: %v", err)
	}
	if err := v.Unmarshal(&config); err != nil {
		log.Fatalf("unmarshal config error: %v", err)
	}
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
	defer rows.Close()
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
			&schedule.ScheduleStatus, &schedule.DisableBetting, &schedule.EnableDisplay)
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
		schedule2.EnableDisplay = schedule.EnableDisplay

		schedules = append(schedules, schedule2)
	}
	return schedules, nil
}

func schedules2(db *sql.DB, scheduleType ScheduleType) ([]Schedule, error) {
	var (
		schedules []Schedule
		rows      *sql.Rows
		err       error
	)

	if scheduleType == All {
		rows, err = db.Query("SELECT * FROM schedule")
	} else {
		rows, err = db.Query("SELECT * FROM schedule WHERE schedule_type = ?", scheduleType)
	}
	if err != nil {
		return []Schedule{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			schedule Schedule
		)
		err := rows.Scan(&schedule.ScheduleID, &schedule.HomeTeam, &schedule.AwayTeam,
			&schedule.HomeTeamWinOdds, &schedule.AwayTeamWinOdds, &schedule.TiedOdds,
			&schedule.ScheduleTime, &schedule.ScheduleGroup, &schedule.ScheduleType,
			&schedule.ScheduleStatus, &schedule.DisableBetting, &schedule.EnableDisplay)
		if err != nil {
			return []Schedule{}, err
		}

		schedules = append(schedules, schedule)
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

func handleSchedules2(c *gin.Context) {
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

	schedules, err := schedules2(db, queryScheduleType)
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

func handleUpdateSchedule(c *gin.Context) {
	var schedule Schedule
	if c.Bind(&schedule) != nil {
		illegalParametersRsp(c)
		return
	}

	// 先验证要更新的赛程是在数据库中
	rows, err := db.Query("SELECT * FROM schedule WHERE schedule_id = ?", schedule.ScheduleID)
	defer rows.Close()
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
				"schedule_status = ?, disable_betting = ?, enable_display = ? " + " WHERE schedule_id = ?")
			if err != nil {
				updateMySQLFailedRsp(c)
				fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
				return
			}
			result, err := stmt.Exec(schedule.HomeTeam, schedule.AwayTeam,
				schedule.HomeTeamWinOdds, schedule.AwayTeamWinOdds, schedule.TiedOdds,
				schedule.ScheduleTime, schedule.ScheduleGroup, schedule.ScheduleType, schedule.ScheduleStatus,
				schedule.DisableBetting, schedule.EnableDisplay, schedule.ScheduleID)
			defer stmt.Close()
			if err != nil {
				operateMySQLFailedRsp(c)
				fmt.Fprintf(os.Stderr, "update schedule failed, result:%v, err: %v\n", result, err)
				return
			}

			// 遍历竞猜表，计算出每个人竞猜的结果，如果竞猜成功，增加相应用户的金币数
			var betRequest BetRequest
			rows, _ := db.Query("SELECT * FROM bet WHERE schedule_id = ? and bet_status = ?",
				schedule.ScheduleID, BetNotFinish)
			for rows.Next() {
				rows.Scan(&betRequest.UserId, &betRequest.ScheduleId, &betRequest.BettingMoney, &betRequest.BettingResult,
					&betRequest.BettingOdds, &betRequest.BettingStatus, &betRequest.WinMoney)
				// 说明竞猜成功
				var (
					money     float64
					win_count int
					win_money float64
					betStatus BetStatus
				)
				if ScheduleStatus(betRequest.BettingResult) == ScheduleStatus(schedule.ScheduleStatus) {
					betStatus = WinBet
					rows, _ := db.Query("SELECT money,win_count FROM user WHERE user_id = ?", betRequest.UserId)
					if rows.Next() {
						rows.Scan(&money, &win_count)
					}
					stmt, _ := db.Prepare("UPDATE user SET money = ?,win_count = ? WHERE user_id = ?")
					win_money = float64(betRequest.BettingMoney) * betRequest.BettingOdds
					currentMoney := win_money + money
					stmt.Exec(currentMoney, win_count+1, betRequest.UserId)
				} else {
					betStatus = LostBet
					win_money = -float64(betRequest.BettingMoney)
				}
				stmt, _ := db.Prepare("UPDATE bet SET bet_status = ?,win_money = ? WHERE user_id = ? and schedule_id = ?")
				stmt.Exec(betStatus, win_money, betRequest.UserId, betRequest.ScheduleId)
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

	rows, err := db.Query("SELECT schedule_id FROM schedule WHERE schedule_time = ? and home_team = ? and away_team = ? ",
		schedule.ScheduleTime, schedule.HomeTeam, schedule.AwayTeam)
	handleError(err)
	defer rows.Close()
	// 如果已经有了这场赛事，就不再插入，避免重复的创建动作
	if rows.Next() {
		var id int
		rows.Scan(&id)
		c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK", "schedule_id": id})
	} else {
		stmt, err := db.Prepare("INSERT INTO " +
			"schedule(home_team,away_team,home_team_win_odds,away_team_win_odds,tied_odds,schedule_time,schedule_group,schedule_type,schedule_status,disable_betting,enable_display) " +
			"VALUES (?,?,?,?,?,?,?,?,?,?,?)")
		if err != nil {
			operateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
			return
		}
		result, err := stmt.Exec(schedule.HomeTeam, schedule.AwayTeam,
			schedule.HomeTeamWinOdds, schedule.AwayTeamWinOdds, schedule.TiedOdds,
			schedule.ScheduleTime, schedule.ScheduleGroup, schedule.ScheduleType,
			schedule.ScheduleStatus, schedule.DisableBetting, schedule.EnableDisplay)
		defer stmt.Close()
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
	handleError(err)
	defer rows.Close()
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
			defer rows.Close()
			if rows.Next() {
				alreadyBet(c)
				return
			}

			var betCount int

			// 验证用户是否有足够的钱进行下注
			rows, err = db.Query("SELECT money, bet_count FROM user WHERE user_id = ?", betRequest.UserId)
			if err != nil {
				operateMySQLFailedRsp(c)
				fmt.Fprintf(os.Stderr, "query mysql failed, err: %v\n", err)
				return
			}
			if rows.Next() {
				var money float64
				err := rows.Scan(&money, &betCount)
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
					"bet(user_id,schedule_id,betting_money,betting_result,betting_odds,bet_status,win_money) " +
					"VALUES (?,?,?,?,?,?,?)")
				defer stmt.Close()
				if err != nil {
					operateMySQLFailedRsp(c)
					fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
					return
				}
				result, err := stmt.Exec(betRequest.UserId, betRequest.ScheduleId,
					betRequest.BettingMoney, betRequest.BettingResult, betRequest.BettingOdds, 0, BetNotFinish)
				if err != nil {
					operateMySQLFailedRsp(c)
					fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
					return
				}

				// 更新 user 表中用户的金币（正确来说，应该这几个操作是一个 事务 操作）
				stmt, err = db.Prepare("UPDATE user SET money = ?, bet_count = ? WHERE user_id = ?")
				if err != nil {
					operateMySQLFailedRsp(c)
					fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
					return
				}
				result, err = stmt.Exec(int(money)-betRequest.BettingMoney, betCount+1, betRequest.UserId)
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
	if config.EnableWhiteList {
		if isIllegalUser(authorizeRequest.ChineseName, authorizeRequest.EnglishName) != true {
			illegalUserRsp(c)
			return
		}
	}

	// 判读是否第一次登陆，如果是第一次登陆，则数据库中找不到相应的记录
	rows, err := db.Query("SELECT * FROM user WHERE chinese_name = ? and rtx_name = ?",
		authorizeRequest.ChineseName, authorizeRequest.EnglishName)
	handleError(err)
	defer rows.Close()
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
		err := rows.Scan(&user.UserId, &user.EnglishName, &user.ChineseName, &user.Password,
			&user.Money, &user.EnableResetPassword, &user.LastLoginTime, &user.WinCount, &user.BetCount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan rows: %v failed, error: %v\n", rows, err)
			queryUserFailedRsp(c)
			return
		}
		if user.Password != authorizeRequest.Password {
			fmt.Fprintf(os.Stderr, "incorrect password, [ch_name:%v, en_name:%v], want: %v, got: %v\n",
				user.ChineseName, user.EnglishName, user.Password, authorizeRequest.Password)
			if user.EnableResetPassword != true {
				incorrectPasswordRsp(c)
				return
			}
		}

		// 更新登陆时间
		if user.EnableResetPassword {
			stmt, err := db.Prepare("UPDATE user SET money = ?, last_login_time = ?, enable_reset_password = ?, password = ? WHERE user_id = ?")
			defer stmt.Close()
			if err != nil {
				updateMySQLFailedRsp(c)
				fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
				return
			}
			result, err := stmt.Exec(user.Money, loginTime, false, authorizeRequest.Password, user.UserId)
			if err != nil {
				updateMySQLFailedRsp(c)
				fmt.Fprintf(os.Stderr, "update schedule failed, result:%v, err: %v\n", result, err)
				return
			}
		} else {
			stmt, err := db.Prepare("UPDATE user SET money = ?, last_login_time = ? WHERE user_id = ?")
			defer stmt.Close()
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
		}

		c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK", "user_id": user.UserId, "money": user.Money})
	} else {
		// 说明是第一次登陆
		stmt, err := db.Prepare("INSERT INTO " +
			"user(rtx_name,chinese_name,password,money,enable_reset_password,last_login_time,win_count,bet_count) " +
			"VALUES (?,?,?,?,?,?,?,?)")
		if err != nil {
			operateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
			return
		}
		result, err := stmt.Exec(authorizeRequest.EnglishName, authorizeRequest.ChineseName, authorizeRequest.Password,
			config.InitialMoney, false, loginTime, 0, 0)
		defer stmt.Close()
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
		tm := time.Unix(loginTimeStamp, 0)
		stmt, err = db.Prepare("INSERT INTO " + "reward(user_id, reward_time, reward_money) " + "VALUES (?,?,?)")
		handleError(err)
		_, err = stmt.Exec(lastId, tm.Format("2006-01-02 15:03:04"), config.InitialMoney)
		handleError(err)
		if err != nil {
			operateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": 0, "desc": "OK", "user_id": lastId, "money": config.InitialMoney, "first_login": true})
	}
}

func isDailyReward(userID int, loginTimeStamp int64) bool {
	tm := time.Unix(loginTimeStamp, 0)
	lowerBound := tm.Format("2006-01-02") + " 00:00:00"
	upperBound := tm.Format("2006-01-02") + " 23:59:59"
	rows, err := db.Query("SELECT * FROM reward WHERE user_id = ? and reward_time > ? and reward_time < ?",
		userID, lowerBound, upperBound)
	handleError(err)
	if err != nil {
		log.Fatalf("query db failed, error: %v\n", err)
	}

	// 如果查到 reward 表中已经有了记录，说明今天已经送过金币
	if rows.Next() {
		return false
	}

	return true
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
	handleError(err)
	defer rows.Close()
	if err != nil {
		log.Fatalf("query db failed, error: %v\n", err)
	}
	for rows.Next() {
		var betRequest BetRequest
		err := rows.Scan(&betRequest.UserId, &betRequest.ScheduleId,
			&betRequest.BettingMoney, &betRequest.BettingResult, &betRequest.BettingOdds, &betRequest.BettingStatus, &betRequest.WinMoney)
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
	handleError(err)
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
			defer stmt.Close()
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
	var grantResetPasswordReq GrantResetPassword
	if c.Bind(&grantResetPasswordReq) != nil {
		illegalParametersRsp(c)
		return
	}

	rows, err := db.Query("SELECT user_id FROM user WHERE rtx_name = ? and chinese_name = ?",
		grantResetPasswordReq.EnglishName, grantResetPasswordReq.ChineseName)
	defer rows.Close()
	handleError(err)
	if rows.Next() {
		var userID int
		err := rows.Scan(&userID)
		handleError(err)
		if err != nil {
			operateMySQLFailedRsp(c)
			return
		}
		stmt, err := db.Prepare("UPDATE user SET enable_reset_password = ? WHERE user_id = ?")
		defer stmt.Close()
		handleError(err)
		if err != nil {
			operateMySQLFailedRsp(c)
			return
		}
		stmt.Exec(true, userID)
		c.JSON(http.StatusOK, gin.H{
			"status": 0,
			"desc":   "OK",
		})
	} else {
		userNotExist(c)
		return
	}
}

func handleRank(c *gin.Context) {
	limit := c.Query("limit")
	if limit == "" {
		limit = "20"
	}

	if DisplayRank {
		if len(GRank) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"status": 0,
				"desc":   "OK",
				"rank":   GRank,
			})
		} else {
			ranks := []RankRsp{}
			rows, err := db.Query("SELECT user_id, rtx_name, chinese_name, money,win_count,bet_count FROM user WHERE bet_count > 0 ORDER BY money desc limit ?", limit)
			handleError(err)
			defer rows.Close()
			if err != nil {
				operateMySQLFailedRsp(c)
				return
			}

			for rows.Next() {
				var rank RankRsp
				err := rows.Scan(&rank.UserID, &rank.RTXName, &rank.ChineseName, &rank.Money, &rank.WinCount, &rank.BetCount)
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
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status": 0,
			"desc":   "OK",
			"rank":   []RankRsp{},
		})
	}
}

func handleRewardHistory(c *gin.Context) {
	userID := c.Query("user_id")

	rewardHistory := []RewardHistory{}
	rows, err := db.Query("SELECT user_id,reward_time,reward_money FROM reward WHERE user_id = ?", userID)
	handleError(err)
	defer rows.Close()
	for rows.Next() {
		var history RewardHistory
		err := rows.Scan(&history.UserId, &history.RewardTime, &history.RewardMoney)
		handleError(err)
		if err != nil {
			operateMySQLFailedRsp(c)
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

func rankNumber(userID int) int {
	rows, err := db.Query("SELECT u.rank FROM (select user_id, (@ranknum:=@ranknum+1) as rank from user,(select (@ranknum :=0) ) b order by money desc)u where u.user_id = ?", userID)
	handleError(err)
	defer rows.Close()
	if rows.Next() {
		var rank int
		rows.Scan(&rank)
		return rank
	}
	return 0
}

func handleMyInfo(c *gin.Context) {
	userID := c.Query("user_id")
	userID2, err := strconv.Atoi(userID)
	if err != nil {
		illegalParametersRsp(c)
		return
	}
	isDailyReward := isDailyReward(userID2, time.Now().Unix())
	rows, err := db.Query("SELECT user_id,rtx_name,chinese_name,money,win_count,bet_count FROM user WHERE user_id = ?", userID2)
	handleError(err)
	defer rows.Close()
	rank := rankNumber(userID2)
	if rows.Next() {
		var user User
		err := rows.Scan(&user.UserId, &user.EnglishName, &user.ChineseName, &user.Money, &user.WinCount, &user.BetCount)
		if err != nil {
			operateMySQLFailedRsp(c)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":       0,
			"desc":         "OK",
			"id":           user.UserId,
			"money":        user.Money,
			"cn_name":      user.ChineseName,
			"en_name":      user.EnglishName,
			"daily_reward": isDailyReward,
			"rank":         rank,
			"win_count":    user.WinCount,
			"bet_count":    user.BetCount,
		})
	} else {
		userNotExist(c)
		return
	}
}

func handleDailyReward(c *gin.Context) {
	var req DailyRewardRequest
	if c.Bind(&req) != nil {
		illegalParametersRsp(c)
		return
	}

	timestamp := time.Now().Unix()
	if isDailyReward(req.UserID, timestamp) {
		tm := time.Unix(timestamp, 0)
		stmt, err := db.Prepare("INSERT INTO " + "reward(user_id, reward_time, reward_money) " + "VALUES (?,?,?)")
		handleError(err)
		_, err = stmt.Exec(req.UserID, tm.Format("2006-01-02 15:03:04"), config.DailyRewardMoney)
		handleError(err)
		defer stmt.Close()
		if err != nil {
			operateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}

		var money float64
		rows, err := db.Query("SELECT money FROM user WHERE user_id = ?", req.UserID)
		handleError(err)
		defer rows.Close()
		if err != nil {
			operateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return
		}
		if rows.Next() {
			err = rows.Scan(&money)
			if err != nil {
				operateMySQLFailedRsp(c)
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return
			}
			stmt, err = db.Prepare("UPDATE user SET money = ? where user_id = ?")
			handleError(err)
			stmt.Exec(money+float64(config.DailyRewardMoney), req.UserID)
			defer stmt.Close()
			c.JSON(http.StatusOK, gin.H{
				"status": 0,
				"desc":   "OK",
			})
		} else {
			userNotExist(c)
			return
		}
	} else {
		alreayGetDailyReward(c)
		return
	}
}

func handleCountry(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  0,
		"desc":    "OK",
		"country": CountryInfoList,
	})
}

func handleTips(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM tips")
	handleError(err)
	defer rows.Close()
	if err != nil {
		operateMySQLFailedRsp(c)
		return
	}
	var (
		tips           Tips
		addTipsRequest AddTipsRequest
	)
	for rows.Next() {
		err := rows.Scan(&tips.TipsID, &tips.Content, &tips.EnableDisplay)
		if err != nil {
			operateMySQLFailedRsp(c)
			return
		}
		addTipsRequest.TipsList = append(addTipsRequest.TipsList, tips)
	}
	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"desc":   "OK",
		"tips":   addTipsRequest,
	})
}

func handleAddTips(c *gin.Context) {
	var addTipsRequest AddTipsRequest
	if c.Bind(&addTipsRequest) != nil {
		illegalParametersRsp(c)
		return
	}

	for _, t := range addTipsRequest.TipsList {
		stmt, err := db.Prepare("REPLACE INTO " +
			"tips(tips_id,content,enable_display) VALUES (?,?,?)")
		if err != nil {
			operateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "sql prepare failed, err: %v\n", err)
			return
		}
		result, err := stmt.Exec(t.TipsID, t.Content, t.EnableDisplay)
		handleError(err)
		defer stmt.Close()
		if err != nil {
			operateMySQLFailedRsp(c)
			fmt.Fprintf(os.Stderr, "insert schedule failed, result:%v, err: %v\n", result, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"desc":   "OK",
	})
}

func handleUploadPictures(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	files := form.File["files"]
	var displayFileList []string
	for _, file := range files {
		if err := c.SaveUploadedFile(file, "./assets/"+file.Filename); err != nil {
			fmt.Fprintf(os.Stderr, "upload file err: %s", err.Error())
			uploadFileFailed(c)
			return
		}
		displayFileList = append(displayFileList, config.DomainName+"/assets/"+file.Filename)
	}

	GdisplayFileList = displayFileList

	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"desc":   fmt.Sprintf("Uploaded successfully %d files", len(files)),
	})
}

func handleDisplay(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  0,
		"desc":    "OK",
		"display": GdisplayFileList,
	})
}

func handleAddNewUser(c *gin.Context) {
	var newUserReq AddNewUserReq
	if c.Bind(&newUserReq) != nil {
		illegalParametersRsp(c)
		return
	}

	if addNewUser(timiNewUsers, newUserReq.ChineseName, newUserReq.EnglishName) != true {
		addNewUserFailed(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"desc":   "OK",
	})
}

func handleUpdateRanks(c *gin.Context) {
	var req UpdateRankReq
	if c.Bind(&req) != nil {
		illegalParametersRsp(c)
		return
	}

	DisplayRank = req.EnableDisplayRank
	GRank = req.Rank

	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"desc":   "OK",
	})
}

func init() {
	parseConfig()
	db = sqlDB()

	readUserFile(config.CSVNameList, config.TimiNewUser)
	if _, err := os.Stat("assets"); os.IsNotExist(err) {
		if err := os.Mkdir("assets", 0755); err != nil {
			log.Fatalf("create assets failed, error: %v\n", err)
		}
	}
}

func handleError(err error) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		log.Printf("[error] %s:%d %v", fn, line, err)
	}
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
	router.GET("/schedules2", handleSchedules2)
	router.GET("/rank", handleRank)
	router.GET("/betting_history", handleBettingHistory)
	router.GET("/reward_history", handleRewardHistory)
	router.GET("/my", handleMyInfo)
	router.GET("/country", handleCountry)
	router.GET("/tips", handleTips)
	router.GET("/display", handleDisplay)

	router.POST("/update_schedule", handleUpdateSchedule)
	router.POST("/bet", handleBet)
	router.POST("/authorize", handleAuthorize)
	router.POST("/daily_reward", handleDailyReward)
	router.POST("/reset_password", handleResetPassword)
	router.POST("/grant_reset_password", handleGrantResetPassword)
	router.POST("/add_tips", handleAddTips)
	router.POST("/upload_pictures", handleUploadPictures)
	router.POST("/add_new_user", handleAddNewUser)
	router.POST("/update_ranks", handleUpdateRanks)

	router.Static("/assets", "./assets")
	router.Run(config.ServerPort)
}
