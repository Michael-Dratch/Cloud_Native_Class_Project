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

type VoterAPI struct {
	db *db.VoterData
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

func (voterAPI *VoterAPI) UpdateVoter(c *gin.Context) {
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

	if id != uint(voter.VoterID) {
		voterAPI.handleBadRequestError(c, "ERROR: ID in url and request body do not match", err)
		return
	}

	err = voterAPI.db.UpdateVoter(voter.VoterID, voter)
	if err != nil {
		voterAPI.handleBadRequestError(c, "Voter does not exist", err)
		return
	}
	c.JSON(http.StatusOK, voter)
}

func (voterAPI *VoterAPI) DeleteVoter(c *gin.Context) {
	voterAPI.totalCalls++
	id, err := getParameterUint(c, "id")
	if err != nil {
		voterAPI.handleBadRequestError(c, "Error converting voter id to int", err)
		return
	}

	err = voterAPI.db.DeleteVoter(id)
	if err != nil {
		voterAPI.handleBadRequestError(c, "Voter does not exist", err)
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
