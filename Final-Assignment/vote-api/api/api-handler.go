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

