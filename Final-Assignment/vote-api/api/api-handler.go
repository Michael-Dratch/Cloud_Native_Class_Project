package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"votes-api/db"

	"github.com/gin-gonic/gin"
)

type VoteAPI struct {
	db *db.VoteData
	bootTime time.Time
	totalCalls int
	totalErrors int
}

type HealthCheckData struct {
	UpTime string
	TotalCalls int
	TotalErrors int
}

func New() (*VoteAPI, error) {
	dbHandler, err := db.New()
	if err != nil {
		return nil, err
	}

	return &VoteAPI{   db: dbHandler, 
						bootTime: time.Now(),
						totalCalls: 0,
						totalErrors: 0,}, nil
}

func (voteAPI *VoteAPI) ListAllVotes(c *gin.Context) {
	voteAPI.totalCalls++
	voterList, err := voteAPI.db.GetAllVotes()
	if err != nil {
		voteAPI.handleInternalServerError(c, "Error Getting All Votes: ", err)
		return
	}

	if voterList == nil {
		voterList = make([]db.Vote, 0)
	}

	c.JSON(http.StatusOK, voterList)
}

func (voteAPI *VoteAPI) GetVote(c *gin.Context) {
	voteAPI.totalCalls++

	id, err := getParameterUint(c, "id")
	if err != nil {
		voteAPI.handleBadRequestError(c, "Error converting vote id to int", err)
		return
	}

	isDetail := c.Query("detail")
	if isDetail == "true"{
		vote, err := voteAPI.db.GetVoteDetails(id)
		if err != nil {
			voteAPI.handleBadRequestError(c, "Vote details not found: ", err)
			return
		}
		c.JSON(http.StatusOK, vote)
	} else {
		vote, err := voteAPI.db.GetVote(id)
		if err != nil {
			voteAPI.handleBadRequestError(c, "Vote not found: ", err)
			return
		}
		c.JSON(http.StatusOK, vote)
	}

	
}

func (voteAPI *VoteAPI) AddVote(c *gin.Context) {
	voteAPI.totalCalls++
	
	id, err := getParameterUint(c, "id")
	if err != nil {
		voteAPI.handleBadRequestError(c, "Error converting voter id to int", err)
		return
	}

	var voteKeys db.VoteKeys
	if err := c.ShouldBindJSON(&voteKeys); err != nil {
		voteAPI.handleBadRequestError(c, "Error binding JSON: ", err)
		return
	}

	if id != uint(voteKeys.VoteID) {
		voteAPI.handleBadRequestError(c, "ERROR: ID in url and request body do not match", err)
		return
	}

	if err := voteAPI.db.AddVote(voteKeys); err != nil {
		voteAPI.handleInternalServerError(c, "Error adding voter: ", err)
		return
	}

	vote,_ := voteAPI.db.GetVote(voteKeys.VoteID)

	c.JSON(http.StatusOK, vote)
}


func (voteAPI *VoteAPI) UpdateVote(c *gin.Context) {
	voteAPI.totalCalls++
	id, err := getParameterUint(c, "id")
	if err != nil {
		voteAPI.handleBadRequestError(c, "Error converting vote id to int", err)
		return
	}

	var voteKeys db.VoteKeys
	if err := c.ShouldBindJSON(&voteKeys); err != nil {
		voteAPI.handleBadRequestError(c, "Error binding JSON: ", err)
		return
	}

	if id != uint(voteKeys.VoteID) {
		voteAPI.handleBadRequestError(c, "ERROR: ID in url and request body do not match", err)
		return
	}

	err = voteAPI.db.UpdateVote(voteKeys.VoteID, voteKeys)
	if err != nil {
		voteAPI.handleBadRequestError(c, "Vote does not exist", err)
		return
	}

	vote,_ := voteAPI.db.GetVote(voteKeys.VoteID)
	c.JSON(http.StatusOK, vote)
}

func (voteAPI *VoteAPI) DeleteVote(c *gin.Context) {
	voteAPI.totalCalls++
	id, err := getParameterUint(c, "id")
	if err != nil {
		voteAPI.handleBadRequestError(c, "Error converting voter id to int", err)
		return
	}

	err = voteAPI.db.DeleteVote(id)
	if err != nil {
		voteAPI.handleBadRequestError(c, "Voter does not exist", err)
		return
	}
	
	c.Status(http.StatusOK)
}

/*
func (voteAPI *VoteAPI) GetVoterHistory(c *gin.Context) {
	voteAPI.totalCalls++

	id, err := getParameterUint(c, "id")
	if err != nil {
		voteAPI.handleBadRequestError(c, "Error converting voter id to int", err)
		return
	}

	voterHistory, err := voteAPI.db.GetVoterHistory(id)
	if err != nil {
		voteAPI.handleBadRequestError(c, "Voter does not exist", err)
		return
	}

	c.JSON(http.StatusOK, voterHistory)
}

func (voteAPI *VoteAPI) GetVoterPoll(c *gin.Context) {
	voteAPI.totalCalls++
	id, err := getParameterUint(c, "id")
	if err != nil { 
		voteAPI.handleBadRequestError(c, "Error converting voter id to int", err)
		return
	}

	pollID, err := getParameterUint(c, "pollid")
	if err != nil { 
		voteAPI.handleBadRequestError(c, "Error converting poll id to int", err)
		return
	}

	poll, err := voteAPI.db.GetVoterPoll(id, pollID)
	if err != nil {
		voteAPI.handleBadRequestError(c, "Voter has not voted in this poll", err)
		return
	}

	c.JSON(http.StatusOK, poll)
}

func (voteAPI *VoteAPI) AddVoterPoll(c *gin.Context) {
	voteAPI.totalCalls++
	voterID, err := getParameterUint(c, "id")
	if err != nil { 
		voteAPI.handleBadRequestError(c, "Error converting voter ID to int", err)
		return
	}

	pollID, err := getParameterUint(c, "pollid")
	if err != nil { 
		voteAPI.handleBadRequestError(c, "Error converting poll id to int", err)
		return
	}

	var voterPoll db.VoterPoll
	if err := c.ShouldBindJSON(&voterPoll); err != nil {
		voteAPI.handleBadRequestError(c, "Error binding JSON: ", err)
		return
	}

	voterPoll.VoteDate = time.Now()

	if pollID != uint(voterPoll.PollID) {
		voteAPI.handleBadRequestError(c, "ERROR: poll ID in url and request body do not match", nil)
		return
	}

	pollExists := voteAPI.db.DoesVoterPollExist(voterID, pollID)
	if pollExists {
		voteAPI.handleBadRequestError(c, "ERROR: Voter cannot vote in the same poll twice: ", err)
		return
	}

	if err := voteAPI.db.AddVoterPoll(voterID, voterPoll); err != nil {
		voteAPI.handleInternalServerError(c, "Error adding voter poll: ", err)
		return
	}

	c.JSON(http.StatusOK, voterPoll)
}

func (voteAPI *VoteAPI) UpdateVoterPoll(c *gin.Context) {
	voteAPI.totalCalls++
	voterID, err := getParameterUint(c, "id")
	if err != nil { 
		voteAPI.handleBadRequestError(c, "Error converting voter ID to int", err)
		return
	}

	pollID, err := getParameterUint(c, "pollid")
	if err != nil { 
		voteAPI.handleBadRequestError(c, "Error converting poll id to int", err)
		return
	}

	pollExists := voteAPI.db.DoesVoterPollExist(voterID, pollID)
	if pollExists == false {
		voteAPI.handleBadRequestError(c, "No vote data exists for this voter in this poll", errors.New("Voter poll does not exist"))
		return
	}

	var voterPoll db.VoterPoll
	if err := c.ShouldBindJSON(&voterPoll); err != nil {
		voteAPI.handleBadRequestError(c, "Error binding JSON: ", err)
		return
	}

	if pollID != uint(voterPoll.PollID) {
		voteAPI.handleBadRequestError(c, "ERROR: poll ID in url and request body do not match", nil)
		return
	}

	voteAPI.db.UpdateVoterPoll(voterID, pollID, voterPoll)
	c.JSON(http.StatusOK, voterPoll)
}

func (voteAPI *VoteAPI) DeleteVoterPoll(c *gin.Context) {
	voteAPI.totalCalls++
	voterID, err := getParameterUint(c, "id")
	if err != nil { 
		voteAPI.handleBadRequestError(c, "Error converting voter ID to int", err)
		return
	}

	pollID, err := getParameterUint(c, "pollid")
	if err != nil { 
		voteAPI.handleBadRequestError(c, "Error converting poll id to int", err)
		return
	}

	pollExists := voteAPI.db.DoesVoterPollExist(voterID, pollID)
	if pollExists == false {
		voteAPI.handleBadRequestError(c, "No vote data exists for this voter in this poll", errors.New("Voter poll does not exist"))
		return
	}

	voteAPI.db.DeleteVoterPoll(voterID, pollID)
	c.Status(http.StatusOK)
}

*/

func getParameterUint(c *gin.Context, name string) (uint, error) {
	paramS := c.Param(name)
	param64, err := strconv.ParseUint(paramS, 10, 64)
	if err != nil {
		return 0, errors.New("Error converting parameter to int64")
	}
	return uint(param64), nil
}

func (voteAPI *VoteAPI) handleBadRequestError(c *gin.Context, errorMessage string, err error) {
	voteAPI.totalErrors++
	log.Println(errorMessage, err)
	c.AbortWithStatus(http.StatusBadRequest)
}

func (voteAPI *VoteAPI) handleInternalServerError(c *gin.Context, errorMessage string, err error) {
	voteAPI.totalErrors++
	log.Println(errorMessage, err)
	c.AbortWithStatus(http.StatusInternalServerError)
}

func (voteAPI * VoteAPI) HealthCheck(c *gin.Context) {
	healthData := HealthCheckData{UpTime: time.Now().Sub(voteAPI.bootTime).String(), 
									TotalCalls: voteAPI.totalCalls,
									TotalErrors: voteAPI.totalErrors,}
	c.IndentedJSON(http.StatusOK, healthData)
}

