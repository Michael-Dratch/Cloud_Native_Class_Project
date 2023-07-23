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
}

func New() (*VoterAPI, error) {
	dbHandler, err := db.New()
	if err != nil {
		return nil, err
	}

	return &VoterAPI{db: dbHandler}, nil
}

func (voterAPI *VoterAPI) ListAllVoters(c *gin.Context) {
	voterList, err := voterAPI.db.GetAllVoters()
	if err != nil {
		log.Println("Error Getting All Voters: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if voterList == nil {
		voterList = make([]db.Voter, 0)
	}

	c.JSON(http.StatusOK, voterList)
}

func (voterAPI *VoterAPI) GetVoter(c *gin.Context) {

	id, err := getParameterUint(c, "id")
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	voter, err := voterAPI.db.GetVoter(id)
	if err != nil {
		log.Println("Voter not found: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, voter)
}

func (voterAPI *VoterAPI) AddVoter(c *gin.Context) {
	
	id, err := getParameterUint(c, "id")
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var voter db.Voter
	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	/* 
	Making the ID a parameter in the url seems redundent and creates the possibility
	of a new error (ID mismatch in the url and the json body). I wonder if including this 
	is valuable because it makes it clear the url pattern is related to a single voter
	or if it could be removed so that the add voter end point would just be a POST call 
	to /voters with the id and other data in the requet body
	*/
	if id != uint(voter.VoterID) {
		log.Println("ERROR: ID in url and request body do not match")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := voterAPI.db.AddVoter(voter); err != nil {
		log.Println("Error adding voter: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, voter)
}

func (voterAPI *VoterAPI) GetVoterHistory(c *gin.Context) {

	id, err := getParameterUint(c, "id")
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	voterHistory, err := voterAPI.db.GetVoterHistory(id)

	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, voterHistory)
}

func (voterAPI *VoterAPI) GetVoterPoll(c *gin.Context) {
	id, err := getParameterUint(c, "id")
	if err != nil { c.AbortWithStatus(http.StatusBadRequest)}

	pollID, err := getParameterUint(c, "pollid")
	if err != nil { c.AbortWithStatus(http.StatusBadRequest)}

	poll, err := voterAPI.db.GetVoterPoll(id, pollID)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusOK, poll)
}

func (voterAPI *VoterAPI) AddVoterPoll(c *gin.Context) {
	
	voterID, err := getParameterUint(c, "id")
	if err != nil { c.AbortWithStatus(http.StatusBadRequest)}

	pollID, err := getParameterUint(c, "pollid")
	if err != nil { c.AbortWithStatus(http.StatusBadRequest)}


	var voterPoll db.VoterPoll
	if err := c.ShouldBindJSON(&voterPoll); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	voterPoll.VoteDate = time.Now()

	if pollID != uint(voterPoll.PollID) {
		log.Println("ERROR: poll ID in url and request body do not match")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}


	if err := voterAPI.db.AddVoterPoll(voterID, voterPoll); err != nil {
		log.Println("Error adding voter poll: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
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




/*

//Below we implement the API functions.  Some of the framework
//things you will see include:
//   1) How to extract a parameter from the URL, for example
//	  the id parameter in /todo/:id
//   2) How to extract the body of a POST request
//   3) How to return JSON and a correctly formed HTTP status code
//	  for example, 200 for OK, 404 for not found, etc.  This is done
//	  using the c.JSON() function
//   4) How to return an error code and abort the request.  This is
//	  done using the c.AbortWithStatus() function

// implementation for GET /todo
// returns all todos
func (td *ToDoAPI) ListAllTodos(c *gin.Context) {

	todoList, err := td.db.GetAllItems()
	if err != nil {
		log.Println("Error Getting All Items: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	//Note that the database returns a nil slice if there are no items
	//in the database.  We need to convert this to an empty slice
	//so that the JSON marshalling works correctly.  We want to return
	//an empty slice, not a nil slice. This will result in the json being []
	if todoList == nil {
		todoList = make([]db.ToDoItem, 0)
	}

	c.JSON(http.StatusOK, todoList)
}

// implementation for GET /v2/todo
// returns todos that are either done or not done
// depending on the value of the done query parameter
// for example, /v2/todo?done=true will return all
// todos that are done.  Note you can have multiple
// query parameters, for example /v2/todo?done=true&foo=bar
func (td *ToDoAPI) ListSelectTodos(c *gin.Context) {
	//lets first load the data
	todoList, err := td.db.GetAllItems()
	if err != nil {
		log.Println("Error Getting Database Items: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	//If the database is empty, make an empty slice so that the
	//JSON marshalling works correctly
	if todoList == nil {
		todoList = make([]db.ToDoItem, 0)
	}

	//Note that the query parameter is a string, so we
	//need to convert it to a bool
	doneS := c.Query("done")

	//if the doneS is empty, then we will return all items
	if doneS == "" {
		c.JSON(http.StatusOK, todoList)
		return
	}

	//Now we can handle the case where doneS is not empty
	//and we need to filter the list based on the doneS value

	done, err := strconv.ParseBool(doneS)
	if err != nil {
		log.Println("Error converting done to bool: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	//Now we need to filter the list based on the done value
	//that was passed in.  We will create a new slice and
	//only add items that match the done value
	var filteredList []db.ToDoItem
	for _, item := range todoList {
		if item.IsDone == done {
			filteredList = append(filteredList, item)
		}
	}

	//Note that the database returns a nil slice if there are no items
	//in the database.  We need to convert this to an empty slice
	//so that the JSON marshalling works correctly.  We want to return
	//an empty slice, not a nil slice. This will result in the json being []
	if filteredList == nil {
		filteredList = make([]db.ToDoItem, 0)
	}

	c.JSON(http.StatusOK, filteredList)
}

// implementation for GET /todo/:id
// returns a single todo
func (td *ToDoAPI) GetToDo(c *gin.Context) {

	//Note go is minimalistic, so we have to get the
	//id parameter using the Param() function, and then
	//convert it to an int64 using the strconv package
	idS := c.Param("id")
	id64, err := strconv.ParseInt(idS, 10, 32)
	if err != nil {
		log.Println("Error converting id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	//Note that ParseInt always returns an int64, so we have to
	//convert it to an int before we can use it.
	todoItem, err := td.db.GetItem(int(id64))
	if err != nil {
		log.Println("Item not found: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	//Git will automatically convert the struct to JSON
	//and set the content-type header to application/json
	c.JSON(http.StatusOK, todoItem)
}

// implementation for POST /todo
// adds a new todo
func (td *ToDoAPI) AddToDo(c *gin.Context) {
	var todoItem db.ToDoItem

	//With HTTP based APIs, a POST request will usually
	//have a body that contains the data to be added
	//to the database.  The body is usually JSON, so
	//we need to bind the JSON to a struct that we
	//can use in our code.
	//This framework exposes the raw body via c.Request.Body
	//but it also provides a helper function ShouldBindJSON()
	//that will extract the body, convert it to JSON and
	//bind it to a struct for us.  It will also report an error
	//if the body is not JSON or if the JSON does not match
	//the struct we are binding to.
	if err := c.ShouldBindJSON(&todoItem); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := td.db.AddItem(todoItem); err != nil {
		log.Println("Error adding item: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, todoItem)
}

// implementation for PUT /todo
// Web api standards use PUT for Updates
func (td *ToDoAPI) UpdateToDo(c *gin.Context) {
	var todoItem db.ToDoItem
	if err := c.ShouldBindJSON(&todoItem); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := td.db.UpdateItem(todoItem); err != nil {
		log.Println("Error updating item: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, todoItem)
}

// implementation for DELETE /todo/:id
// deletes a todo
func (td *ToDoAPI) DeleteToDo(c *gin.Context) {
	idS := c.Param("id")
	id64, _ := strconv.ParseInt(idS, 10, 32)

	if err := td.db.DeleteItem(int(id64)); err != nil {
		log.Println("Error deleting item: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// implementation for DELETE /todo
// deletes all todos
func (td *ToDoAPI) DeleteAllToDo(c *gin.Context) {

	if err := td.db.DeleteAll(); err != nil {
		log.Println("Error deleting all items: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

/*   SPECIAL HANDLERS FOR DEMONSTRATION - CRASH SIMULATION AND HEALTH CHECK */

// implementation for GET /crash
// This simulates a crash to show some of the benefits of the
// gin framework

/*
func (td *ToDoAPI) CrashSim(c *gin.Context) {
	//panic() is go's version of throwing an exception
	panic("Simulating an unexpected crash")
}

// implementation of GET /health. It is a good practice to build in a
// health check for your API.  Below the results are just hard coded
// but in a real API you can provide detailed information about the
// health of your API with a Health Check
func (td *ToDoAPI) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK,
		gin.H{
			"status":             "ok",
			"version":            "1.0.0",
			"uptime":             100,
			"users_processed":    1000,
			"errors_encountered": 10,
		})
}

*/