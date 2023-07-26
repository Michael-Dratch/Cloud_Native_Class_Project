package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"voter-api/db"

	"github.com/gin-gonic/gin"
)

/*
	r.GET("/voters", apiHandler.ListAllVoters)
	r.GET("/voters/:id", apiHandler.GetVoter)
	r.POST("/voters/:id", apiHandler.AddVoter)
	r.GET("/voters/:id/polls", apiHandler.GetVoterHistory)
	r.GET("/voters/:id/polls/:pollid", apiHandler.GetVoterPoll)
	r.POST("/voters/:id/polls/:pollid", apiHandler.AddVoterPoll)
	r.GET("/voters/health", apiHandler.HealthCheck)
*/
// The api package creates and maintains a reference to the data handler
// this is a good design practice
type VoterAPI struct {
	db *db.VoterList
	bootTime time.Time
	totalCalls int
	totalErrors int
}

type HealthCheckData struct {
	UpTime string
	TotalCalls int
	TotalErrors int
}

func New() (*VoterAPI, error) {
	dbHandler, err := db.New()
	if err != nil {
		return nil, err
	}

	return &VoterAPI{   db: dbHandler, 
						bootTime: time.Now(),
						totalCalls: 0,
						totalErrors: 0,}, nil
}

func (voterAPI *VoterAPI) ListAllVoters(c *gin.Context) {
	voterAPI.totalCalls++
	voterList, err := voterAPI.db.GetAllVoters()
	if err != nil {
		voterAPI.handleInternalServerError(c, "Error Getting All Voters: ", err)
		return
	}

	if voterList == nil {
		voterList = make([]db.Voter, 0)
	}

	c.JSON(http.StatusOK, voterList)
}

func (voterAPI *VoterAPI) GetVoter(c *gin.Context) {
	voterAPI.totalCalls++
	id, err := getParameterUint(c, "id")
	if err != nil {
		voterAPI.handleBadRequestError(c, "Error converting voter id to int", err)
		return
	}

	voter, err := voterAPI.db.GetVoter(id)
	if err != nil {
		voterAPI.handleBadRequestError(c, "Voter not found: ", err)
		return
	}

	c.JSON(http.StatusOK, voter)
}

func (voterAPI *VoterAPI) AddVoter(c *gin.Context) {
	voterAPI.totalCalls++
	
	id, err := getParameterUint(c, "id")
	if err != nil {
		voterAPI.handleBadRequestError(c, "Error converting voter id to int", err)
		return
	}

	var voter db.Voter
	if err := c.ShouldBindJSON(&voter); err != nil {
		voterAPI.handleBadRequestError(c, "Error binding JSON: ", err)
		return
	}

	/* 
	Making the ID a parameter in the url seems redundent and creates the possibility
	of a new error (ID mismatch in the url and the json body). I wonder if including this 
	is valuable because it makes it clear the url pattern is related to a single voter
	or could be removed so that the add voter end point would just be a POST call 
	to /voters with the id and other data in the requet body
	*/
	if id != uint(voter.VoterID) {
		voterAPI.handleBadRequestError(c, "ERROR: ID in url and request body do not match", err)
		return
	}

	if err := voterAPI.db.AddVoter(voter); err != nil {
		voterAPI.handleInternalServerError(c, "Error adding voter: ", err)
		return
	}

	c.JSON(http.StatusOK, voter)
}

func (voterAPI *VoterAPI) GetVoterHistory(c *gin.Context) {
	voterAPI.totalCalls++

	id, err := getParameterUint(c, "id")
	if err != nil {
		voterAPI.handleBadRequestError(c, "Error converting voter id to int", err)
		return
	}

	voterHistory, err := voterAPI.db.GetVoterHistory(id)
	if err != nil {
		voterAPI.handleBadRequestError(c, "Voter does not exist", err)
		return
	}

	c.JSON(http.StatusOK, voterHistory)
}

func (voterAPI *VoterAPI) GetVoterPoll(c *gin.Context) {
	voterAPI.totalCalls++
	id, err := getParameterUint(c, "id")
	if err != nil { 
		voterAPI.handleBadRequestError(c, "Error converting voter id to int", err)
		return
	}

	pollID, err := getParameterUint(c, "pollid")
	if err != nil { 
		voterAPI.handleBadRequestError(c, "Error converting poll id to int", err)
		return
	}

	poll, err := voterAPI.db.GetVoterPoll(id, pollID)
	if err != nil {
		voterAPI.handleBadRequestError(c, "Voter has not voted in this poll", err)
		return
	}
	c.JSON(http.StatusOK, poll)
}

func (voterAPI *VoterAPI) AddVoterPoll(c *gin.Context) {
	voterAPI.totalCalls++
	voterID, err := getParameterUint(c, "id")
	if err != nil { 
		voterAPI.handleBadRequestError(c, "Error converting voter ID to int", err)
		return
	}

	pollID, err := getParameterUint(c, "pollid")
	if err != nil { 
		voterAPI.handleBadRequestError(c, "Error converting poll id to int", err)
		return
	}


	var voterPoll db.VoterPoll
	if err := c.ShouldBindJSON(&voterPoll); err != nil {
		voterAPI.handleBadRequestError(c, "Error binding JSON: ", err)
		return
	}

	voterPoll.VoteDate = time.Now()

	if pollID != uint(voterPoll.PollID) {
		voterAPI.handleBadRequestError(c, "ERROR: poll ID in url and request body do not match", nil)
		return
	}

	pollExists := voterAPI.db.DoesVoterPollExist(voterID, pollID)
	if pollExists {
		voterAPI.handleBadRequestError(c, "ERROR: Voter cannot vote in the same poll twice: ", err)
		return
	}


	if err := voterAPI.db.AddVoterPoll(voterID, voterPoll); err != nil {
		voterAPI.handleInternalServerError(c, "Error adding voter poll: ", err)
		return
	}

	c.JSON(http.StatusOK, voterPoll)
}


func getParameterUint(c *gin.Context, name string) (uint, error) {
	paramS := c.Param(name)
	param64, err := strconv.ParseUint(paramS, 10, 64)
	if err != nil {
		return 0, errors.New("Error converting parameter to int64")
	}
	return uint(param64), nil
}

func (voterAPI *VoterAPI) handleBadRequestError(c *gin.Context, errorMessage string, err error) {
	voterAPI.totalErrors++
	log.Println(errorMessage, err)
	c.AbortWithStatus(http.StatusBadRequest)
}

func (voterAPI *VoterAPI) handleInternalServerError(c *gin.Context, errorMessage string, err error) {
	voterAPI.totalErrors++
	log.Println(errorMessage, err)
	c.AbortWithStatus(http.StatusInternalServerError)
}

func (voterAPI * VoterAPI) HealthCheck(c *gin.Context) {
	healthData := HealthCheckData{UpTime: time.Now().Sub(voterAPI.bootTime).String(), 
									TotalCalls: voterAPI.totalCalls,
									TotalErrors: voterAPI.totalErrors,}
	c.IndentedJSON(http.StatusOK, healthData)
}
