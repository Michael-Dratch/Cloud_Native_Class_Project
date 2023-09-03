package main

import (
	"flag"
	"fmt"
	"os"

	"poll-api/api"

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

	r.GET("polls/", apiHandler.ListAllPolls)
	
	r.GET("polls/:id", apiHandler.GetPoll)
	r.POST("polls/:id", apiHandler.AddPoll)
	r.PUT("polls/:id", apiHandler.UpdatePoll)
	r.DELETE("polls/:id", apiHandler.DeletePoll)

	r.GET("polls/:id/polloption/:optionid", apiHandler.GetPollOption)
	r.POST("polls/:id/polloption/:optionid", apiHandler.AddPollOption)
	r.PUT("polls/:id/polloption/:optionid", apiHandler.UpdatePollOption)
	r.DELETE("polls/:id/polloption/:optionid", apiHandler.DeletePollOption)

	r.GET("/voters/health", apiHandler.HealthCheck)
	
	serverPath := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	r.Run(serverPath)
}