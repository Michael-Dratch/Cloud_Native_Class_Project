package main

import (
	"errors"
	"fmt"
	"os"

	"drexel.edu/todo/db"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Global variables to hold the command line flags to drive the todo CLI
// application
var (
	dbFileNameFlag string
	listFlag       bool
	itemStatusFlag bool
	queryFlag      int
	addFlag        string
	updateFlag     string
	deleteFlag     int
)
type AppOptType int

// To make the code a little more clean we will use the following
// constants as basically enumerations for different options.  This
// allows us to use a switch statement in main to process the command line
// flags effectively
const (
	LIST_DB_ITEM AppOptType = iota
	QUERY_DB_ITEM
	ADD_DB_ITEM
	UPDATE_DB_ITEM
	DELETE_DB_ITEM
	CHANGE_ITEM_STATUS
	NOT_IMPLEMENTED
	INVALID_APP_OPT
)

var rootCmd = &cobra.Command{
	Use: "ToDo",
	Short: "A command line todo list application",
	Run: func(cmd *cobra.Command, args []string){
		opts, err := processCmdLineFlags(cmd)
		if err!= nil {
			fmt.Println(err)
			os.Exit(1)
		}
		processOptions(opts)
	},
}

func init() {
	rootCmd.Flags().StringVar(&dbFileNameFlag, "db", "./data/todo.json", "Name of the database file")
	rootCmd.Flags().BoolVarP(&listFlag, "l", "l", false, "List all the items in the database")
	rootCmd.Flags().IntVarP(&queryFlag, "q","q", 0, "Query an item in the database")
	rootCmd.Flags().StringVarP(&addFlag, "a","a", "", "Add an item to the database")
	rootCmd.Flags().StringVarP(&updateFlag, "u","u", "", "Update an item in the database")
	rootCmd.Flags().IntVarP(&deleteFlag, "d","d", 0, "Delete an item from the database")
	rootCmd.Flags().BoolVarP(&itemStatusFlag, "s","s", false, "Change item 'done' status to true or false")
}

// processCmdLineFlags parses the command line flags for our CLI

//			 REQUIRED:     Study the code below, and make sure you understand
//						   how it works.  Go online and readup on how the
//						   flag package works.  Then, write a nice comment
//				  		   block to document this function that highights that
//						   you understand how it works.
//
//			 EXTRA CREDIT: The best CLI and command line processor for
//						   go is called Cobra.  Refactor this function to
//						   use it.  See github.com/spf13/cobra for information
//						   on how to use it.
//
//	 YOUR ANSWER: 
/*
		This function first defines the valid flags that the program accepts.
		It uses the flag Var function corrosponding to the type of each flag
		to define the variable that will accept the flag value, the flag name, 
		initial value, and usage message. Then the flags are parsed and the function 
		checks to make sure that at least one valid flag has been set.
		Finally, the function uses a switch statement to check the type of the flag
		and set the appOpt variable to the correct operation value defined above,
		which is then returned. 
*/
func processCmdLineFlags(cmd *cobra.Command) (AppOptType, error) {
	var appOpt AppOptType = INVALID_APP_OPT

	//show help if no flags are set
	if len(os.Args) == 1 {
		fmt.Println("Flags: ")
		printFlagUsage(cmd)
	
		return appOpt, errors.New("no flags were set")
	}

	// Loop over the flags and check which ones are set, set appOpt
	// accordingly
	cmd.Flags().Visit(func(f *pflag.Flag) {
		switch f.Name {
		case "l":
			appOpt = LIST_DB_ITEM
		case "q":
			appOpt = QUERY_DB_ITEM
		case "a":
			appOpt = ADD_DB_ITEM
		case "u":
			appOpt = UPDATE_DB_ITEM
		case "d":
			appOpt = DELETE_DB_ITEM
		case "s":
			if appOpt == QUERY_DB_ITEM{
				appOpt = CHANGE_ITEM_STATUS
			} else {
				fmt.Println("Must include query flag (-q) with item id in order to change item status (-s flag)")
				appOpt = INVALID_APP_OPT
			}
		default:
			appOpt = INVALID_APP_OPT
		}
	})

	if appOpt == INVALID_APP_OPT || appOpt == NOT_IMPLEMENTED {
		fmt.Println("Invalid option set or the desired option is not currently implemented")
		printFlagUsage(cmd)
		return appOpt, errors.New("no flags or unimplemented were set")
	}

	return appOpt, nil
}

func printFlagUsage(cmd *cobra.Command){
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "db"{
			fmt.Println("-- "+f.Name + " " + f.Usage)
		} else{
			fmt.Println("- "+f.Name + " " + f.Usage)
		}
	})
}

func processOptions(opts AppOptType){
	//Create a new db object
	todo, err := db.New(dbFileNameFlag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//Switch over the command line flags and call the appropriate
	//function in the db package
	switch opts {
	case LIST_DB_ITEM:
		fmt.Println("Running QUERY_DB_ITEM...")
		todoList, err := todo.GetAllItems()
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
		for _, item := range todoList {
			todo.PrintItem(item)
		}
		fmt.Println("THERE ARE", len(todoList), "ITEMS IN THE DB")
		fmt.Println("Ok")

	case QUERY_DB_ITEM:
		fmt.Println("Running QUERY_DB_ITEM...")
		item, err := todo.GetItem(queryFlag)
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
		todo.PrintItem(item)
		fmt.Println("Ok")
	case ADD_DB_ITEM:
		fmt.Println("Running ADD_DB_ITEM...")
		item, err := todo.JsonToItem(addFlag)
		if err != nil {
			fmt.Println("Add option requires a valid JSON todo item string")
			fmt.Println("Error: ", err)
			break
		}
		if err := todo.AddItem(item); err != nil {
			fmt.Println("Error: ", err)
			break
		}
		fmt.Println("Ok")
	case UPDATE_DB_ITEM:
		fmt.Println("Running UPDATE_DB_ITEM...")
		item, err := todo.JsonToItem(updateFlag)
		if err != nil {
			fmt.Println("Update option requires a valid JSON todo item string")
			fmt.Println("Error: ", err)
			break
		}
		if err := todo.UpdateItem(item); err != nil {
			fmt.Println("Error: ", err)
			break
		}
		fmt.Println("Ok")
	case DELETE_DB_ITEM:
		fmt.Println("Running DELETE_DB_ITEM...")
		err := todo.DeleteItem(deleteFlag)
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
		fmt.Println("Ok")
	case CHANGE_ITEM_STATUS:
		fmt.Println("Running CHANGE_ITEM_STATUS...")
		item, err := todo.GetItem(queryFlag)
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
		item.IsDone = itemStatusFlag
		err = todo.UpdateItem(item)
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
		todo.PrintItem(item)
		fmt.Println("Ok")
	default:
		fmt.Println("INVALID_APP_OPT")
	}
}

// main is the entry point for our todo CLI application.  It processes
// the command line flags and then uses the db package to perform the
// requested operation
func main() {
	Execute()
}

func Execute() error {
	return rootCmd.Execute()
}