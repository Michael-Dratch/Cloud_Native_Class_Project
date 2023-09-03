package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"poll-api/db"

	"github.com/gin-gonic/gin"
)

type PollAPI struct {
	db *db.PollData
	bootTime time.Time
	totalCalls int
	totalErrors int
}

type HealthCheckData struct {
	UpTime string
	TotalCalls int
	TotalErrors int
}

func New() (*PollAPI, error) {
	dbHandler, err := db.New()
	if err != nil {
		return nil, err
	}

	return &PollAPI{   db: dbHandler, 
						bootTime: time.Now(),
						totalCalls: 0,
						totalErrors: 0,}, nil
}

func (pollAPI *PollAPI) ListAllPolls(c *gin.Context) {
	pollAPI.totalCalls++
	pollList, err := pollAPI.db.GetAllPolls()
	if err != nil {
		pollAPI.handleInternalServerError(c, "Error Getting All Polls: ", err)
		return
	}

	if pollList == nil {
		pollList = make([]db.Poll, 0)
	}

	c.JSON(http.StatusOK, pollList)
}

func (pollAPI *PollAPI) GetPoll(c *gin.Context) {
	pollAPI.totalCalls++
	id, err := getParameterUint(c, "id")
	if err != nil {
		pollAPI.handleBadRequestError(c, "Error converting poll id to int", err)
		return
	}

	poll, err := pollAPI.db.GetPoll(id)
	if err != nil {
		pollAPI.handleBadRequestError(c, "Poll not found: ", err)
		return
	}

	c.JSON(http.StatusOK, poll)
}

func (pollAPI *PollAPI) AddPoll(c *gin.Context) {
	pollAPI.totalCalls++
	
	id, err := getParameterUint(c, "id")
	if err != nil {
		pollAPI.handleBadRequestError(c, "Error converting poll id to int", err)
		return
	}

	var poll db.Poll
	if err := c.ShouldBindJSON(&poll); err != nil {
		pollAPI.handleBadRequestError(c, "Error binding JSON: ", err)
		return
	}

	if id != uint(poll.PollID) {
		pollAPI.handleBadRequestError(c, "ERROR: ID in url and request body do not match", err)
		return
	}

	if err := pollAPI.db.AddPoll(poll); err != nil {
		pollAPI.handleInternalServerError(c, "Error adding poll: ", err)
		return
	}

	c.JSON(http.StatusOK, poll)
}

func (pollAPI *PollAPI) UpdatePoll(c *gin.Context) {
	pollAPI.totalCalls++
	id, err := getParameterUint(c, "id")
	if err != nil {
		pollAPI.handleBadRequestError(c, "Error converting poll id to int", err)
		return
	}

	var poll db.Poll
	if err := c.ShouldBindJSON(&poll); err != nil {
		pollAPI.handleBadRequestError(c, "Error binding JSON: ", err)
		return
	}

	if id != uint(poll.PollID) {
		pollAPI.handleBadRequestError(c, "ERROR: ID in url and request body do not match", err)
		return
	}

	err = pollAPI.db.UpdatePoll(poll.PollID, poll)
	if err != nil {
		pollAPI.handleBadRequestError(c, "Poll does not exist", err)
		return
	}
	c.JSON(http.StatusOK, poll)
}

func (pollAPI *PollAPI) DeletePoll(c *gin.Context) {
	pollAPI.totalCalls++
	id, err := getParameterUint(c, "id")
	if err != nil {
		pollAPI.handleBadRequestError(c, "Error converting poll id to int", err)
		return
	}

	err = pollAPI.db.DeletePoll(id)
	if err != nil {
		pollAPI.handleBadRequestError(c, "Poll does not exist", err)
		return
	}
	
	c.Status(http.StatusOK)
}

func (pollAPI *PollAPI) GetPollOptions(c *gin.Context) {
	pollAPI.totalCalls++

	pollID, err := getParameterUint(c, "id")
	if err != nil {
		pollAPI.handleBadRequestError(c, "Error converting poll id to int", err)
		return
	}

	pollOptions, err := pollAPI.db.GetPollOptions(pollID)
	if err != nil {
		pollAPI.handleBadRequestError(c, "Poll does not exist", err)
		return
	}

	c.JSON(http.StatusOK, pollOptions)
}

func (pollAPI *PollAPI) GetPollOption(c *gin.Context) {
	pollAPI.totalCalls++
	pollID, err := getParameterUint(c, "id")
	if err != nil { 
		pollAPI.handleBadRequestError(c, "Error converting poll id to int", err)
		return
	}

	optionID, err := getParameterUint(c, "optionid")
	if err != nil { 
		pollAPI.handleBadRequestError(c, "Error converting poll option id to int", err)
		return
	}

	poll, err := pollAPI.db.GetPollOption(pollID, optionID)
	if err != nil {
		pollAPI.handleBadRequestError(c, "Poll does not have this option", err)
		return
	}

	c.JSON(http.StatusOK, poll)
}

func (pollAPI *PollAPI) AddPollOption(c *gin.Context) {
	pollAPI.totalCalls++
	pollID, err := getParameterUint(c, "id")
	if err != nil { 
		pollAPI.handleBadRequestError(c, "Error converting poll ID to int", err)
		return
	}

	optionID, err := getParameterUint(c, "optionid")
	if err != nil { 
		pollAPI.handleBadRequestError(c, "Error converting poll option id to int", err)
		return
	}

	var pollOption db.PollOption
	if err := c.ShouldBindJSON(&pollOption); err != nil {
		pollAPI.handleBadRequestError(c, "Error binding JSON: ", err)
		return
	}

	if optionID != uint(pollOption.PollOptionID) {
		pollAPI.handleBadRequestError(c, "ERROR: poll option ID in url and request body do not match", nil)
		return
	}

	pollExists := pollAPI.db.DoesPollOptionExist(pollID, optionID)
	if pollExists {
		pollAPI.handleBadRequestError(c, "ERROR: Poll Option ID already exists in Poll ", err)
		return
	}

	if err := pollAPI.db.AddPollOption(pollID, pollOption); err != nil {
		pollAPI.handleInternalServerError(c, "Error adding poll option: ", err)
		return
	}

	c.JSON(http.StatusOK, pollOption)
}

func (pollAPI *PollAPI) UpdatePollOption(c *gin.Context) {
	pollAPI.totalCalls++
	pollID, err := getParameterUint(c, "id")
	if err != nil { 
		pollAPI.handleBadRequestError(c, "Error converting poll ID to int", err)
		return
	}

	optionID, err := getParameterUint(c, "optionid")
	if err != nil { 
		pollAPI.handleBadRequestError(c, "Error converting poll id to int", err)
		return
	}

	pollExists := pollAPI.db.DoesPollOptionExist(pollID, optionID)
	if pollExists == false {
		pollAPI.handleBadRequestError(c, "No option exists for this poll option ID in this poll", errors.New("Poll option does not exist"))
		return
	}

	var pollOption db.PollOption
	if err := c.ShouldBindJSON(&pollOption); err != nil {
		pollAPI.handleBadRequestError(c, "Error binding JSON: ", err)
		return
	}

	if optionID != uint(pollOption.PollOptionID) {
		pollAPI.handleBadRequestError(c, "ERROR: poll option ID in url and request body do not match", nil)
		return
	}

	pollAPI.db.UpdatePollOption(pollID, optionID, pollOption)
	c.JSON(http.StatusOK, pollOption)
}

func (pollAPI *PollAPI) DeletePollOption(c *gin.Context) {
	pollAPI.totalCalls++
	pollID, err := getParameterUint(c, "id")
	if err != nil { 
		pollAPI.handleBadRequestError(c, "Error converting poll ID to int", err)
		return
	}

	optionID, err := getParameterUint(c, "optionid")
	if err != nil { 
		pollAPI.handleBadRequestError(c, "Error converting poll option id to int", err)
		return
	}

	pollOptionExists := pollAPI.db.DoesPollOptionExist(pollID, optionID)
	if pollOptionExists == false {
		pollAPI.handleBadRequestError(c, "No poll option data exists for this poll option in this poll", errors.New("Pollpoll does not exist"))
		return
	}

	pollAPI.db.DeletePollOption(pollID, optionID)
	c.Status(http.StatusOK)
}

func getParameterUint(c *gin.Context, name string) (uint, error) {
	paramS := c.Param(name)
	param64, err := strconv.ParseUint(paramS, 10, 64)
	if err != nil {
		return 0, errors.New("Error converting parameter to int64")
	}
	return uint(param64), nil
}

func (pollAPI *PollAPI) handleBadRequestError(c *gin.Context, errorMessage string, err error) {
	pollAPI.totalErrors++
	log.Println(errorMessage, err)
	c.AbortWithStatus(http.StatusBadRequest)
}

func (pollAPI *PollAPI) handleInternalServerError(c *gin.Context, errorMessage string, err error) {
	pollAPI.totalErrors++
	log.Println(errorMessage, err)
	c.AbortWithStatus(http.StatusInternalServerError)
}

func (pollAPI *PollAPI) HealthCheck(c *gin.Context) {
	healthData := HealthCheckData{UpTime: time.Now().Sub(pollAPI.bootTime).String(), 
									TotalCalls: pollAPI.totalCalls,
									TotalErrors: pollAPI.totalErrors,}
	c.IndentedJSON(http.StatusOK, healthData)
}
