package main

import (
	"flag"
	"fmt"
	"os"

	"voter-api/api"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Global variables to hold the command line flags to drive the todo CLI
// application
var (
	hostFlag string
	portFlag uint
)

func processCmdLineFlags() {

	flag.StringVar(&hostFlag, "h", "0.0.0.0", "Listen on all interfaces")
	flag.UintVar(&portFlag, "p", 1080, "Default Port")

	flag.Parse()
}

func main() {
	processCmdLineFlags()
	r := gin.Default()
	r.Use(cors.Default())

	apiHandler, err := api.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	r.GET("polls/")
	
	r.GET("polls/:pollid", apiHandler.GetVoterPoll)
	r.POST("polls/:pollid", apiHandler.AddVoterPoll)
	r.PUT("polls/:pollid", apiHandler.UpdateVoterPoll)
	r.DELETE("polls/:pollid", apiHandler.DeleteVoterPoll)

	r.GET("polls/:pollid/polloption/:polloptionid", apiHandler.GetVoterPoll)
	r.POST("polls/:pollid/polloption/:polloptionid", apiHandler.AddVoterPoll)
	r.PUT("polls/:pollid/polloption/:polloptionid", apiHandler.UpdateVoterPoll)
	r.DELETE("polls/:pollid/polloption/:polloptionid", apiHandler.DeleteVoterPoll)
	
	r.GET("/voters/health", apiHandler.HealthCheck)
	
	serverPath := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	r.Run(serverPath)
}