package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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

func queryMySQLFailedRsp(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"status": 3,
		"desc":   "Query mysql failed",
	})
}

func updateMySQLFailedRsp(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"status": 4,
		"desc":   "Update mysql failed",
	})
}

func operateMySQLFailedRsp(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"status": 5,
		"desc":   "Operate mysql failed",
	})
}

func illegalUserRsp(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status": 6,
		"desc":   "Illegal User",
	})
}

func queryUserFailedRsp(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status": 7,
		"desc":   "Query user failed",
	})
}

func incorrectPasswordRsp(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status": 8,
		"desc":   "Incorrect password",
	})
}

func userNotExist(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status": 9,
		"desc":   "User is not exist",
	})
}

func alreadyBet(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status": 10,
		"desc":   "Already bet",
	})
}

func notEnoughMoney(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status": 11,
		"desc":   "Not enough money",
	})
}

func notAllowResetPassword(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status": 12,
		"desc":   "Not allow to reset password",
	})
}

func disableBet(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status": 12,
		"desc":   "Disable bet",
	})
}

func alreayGetDailyReward(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status": 13,
		"desc":   "Already get daily reward",
	})
}
