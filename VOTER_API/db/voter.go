package db

import (
	"errors"
	"log"
	"time"
)

type VoterPoll struct {
	PollID uint
	VoteDate time.Time
}

type Voter struct {
	VoterID uint
	FirstName string
	LastName string
	VoteHistory []VoterPoll
}

type VoterList struct {
	Voters map[uint]Voter
}

// Creat New Voter Data Handler 
func New() (*VoterList, error){
	voterList := &VoterList{
		Voters: make(map[uint]Voter),
	}

	history := make ([]VoterPoll, 0)
	history = append(history, VoterPoll{PollID:1, VoteDate: time.Now()})
	voterList.Voters[1] = Voter{
		VoterID: 1,
		FirstName: "Michael",
		LastName: "Dratch",
		VoteHistory: history,
	}

	return voterList, nil
}

func NewVoter(voterID uint, firstName string, lastName string) (*Voter, error){
	voter := &Voter{
		VoterID: voterID,
		FirstName: firstName,
		LastName: lastName,
		VoteHistory: make ([]VoterPoll, 0),
	}

	return voter, nil
}

func NewVoterPoll(pollID uint) (*VoterPoll, error){
	voterPoll := &VoterPoll{
		PollID: pollID,
		VoteDate: time.Now(),
	}

	return voterPoll, nil
}


func (v *VoterList) GetAllVoters() ([]Voter, error){
	var voters []Voter
	for _, item := range v.Voters {
		voters = append(voters, item)
	}

	return voters, nil
} 

func (v *VoterList) GetVoter(voterID uint) (Voter, error){
	
	voter, ok := v.Voters[voterID]
	if ok {
		return voter, nil
	} else {
		return Voter{}, errors.New("Voter ID does not exist")
	}
} 

func (v *VoterList) AddVoter(voter Voter) error {
	_, ok := v.Voters[voter.VoterID]
	if ok {
		return errors.New("Voter already exists")
	}

	v.Voters[voter.VoterID] = voter

	return nil
}

func (v *VoterList) GetVoterHistory(voterID uint) ([]VoterPoll, error){
	voter, err := v.GetVoter(voterID)
	if err != nil {
		return make([]VoterPoll, 0) , errors.New("Voter ID does not exist")
	}
	
	return voter.VoteHistory, nil
	
}

func (v *VoterList) GetVoterPoll(voterID uint, pollID uint) (VoterPoll, error){
	voter, err := v.GetVoter(voterID)
	if err != nil {
		return VoterPoll{} , errors.New("Voter ID does not exist")
	} 

	for _, poll := range voter.VoteHistory{
		if poll.PollID == pollID{
			return poll, nil
		}
	}
	log.Println("Error: Poll ID does not exist for this voter")
	return VoterPoll{}, errors.New("Poll ID does not exist for this voter")
}

func (v *VoterList) DoesVoterPollExist(voterID uint, pollID uint) bool {
	_, err := v.GetVoterPoll(voterID, pollID)
	if err == nil { return true } 
	return false 
}

func (v *VoterList) AddVoterPoll(voterID uint, newPoll VoterPoll) error{
	voter, err := v.GetVoter(voterID)
	if err != nil {
		return errors.New("Voter ID does not exist")
	} 

	for _, poll := range voter.VoteHistory{
		if newPoll.PollID == poll.PollID{
			return errors.New("Poll ID already exists for this voter")
		}
	}

	voter.VoteHistory = append(voter.VoteHistory, newPoll)
	v.Voters[voterID] = voter
	return nil
}


/*
func (t *ToDo) AddItem(item ToDoItem) error {

	//Before we add an item to the DB, lets make sure
	//it does not exist, if it does, return an error
	_, ok := t.toDoMap[item.Id]
	if ok {
		return errors.New("item already exists")
	}

	//Now that we know the item doesn't exist, lets add it to our map
	t.toDoMap[item.Id] = item

	//If everything is ok, return nil for the error
	return nil
}

// DeleteItem accepts an item id and removes it from the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The item must exist in the DB
//	    				because we use the item.Id as the key, this
//						function must check if the item already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	 (1) The item will be removed from the DB
//		(2) The DB file will be saved with the item removed
//		(3) If there is an error, it will be returned
func (t *ToDo) DeleteItem(id int) error {

	// we should if item exists before trying to delete it
	// this is a good practice, return an error if the
	// item does not exist

	//Now lets use the built-in go delete() function to remove
	//the item from our map
	delete(t.toDoMap, id)

	return nil
}

// DeleteAll removes all items from the DB.
// It will be exposed via a DELETE /todo endpoint
func (t *ToDo) DeleteAll() error {
	//To delete everything, we can just create a new map
	//and assign it to our existing map.  The garbage collector
	//will clean up the old map for us
	t.toDoMap = make(map[int]ToDoItem)

	return nil
}

// UpdateItem accepts a ToDoItem and updates it in the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The item must exist in the DB
//	    				because we use the item.Id as the key, this
//						function must check if the item already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	 (1) The item will be updated in the DB
//		(2) The DB file will be saved with the item updated
//		(3) If there is an error, it will be returned
func (t *ToDo) UpdateItem(item ToDoItem) error {

	// Check if item exists before trying to update it
	// this is a good practice, return an error if the
	// item does not exist
	_, ok := t.toDoMap[item.Id]
	if !ok {
		return errors.New("item does not exist")
	}

	//Now that we know the item exists, lets update it
	t.toDoMap[item.Id] = item

	return nil
}

// GetItem accepts an item id and returns the item from the DB.
// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The item must exist in the DB
//	    				because we use the item.Id as the key, this
//						function must check if the item already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	 (1) The item will be returned, if it exists
//		(2) If there is an error, it will be returned
//			along with an empty ToDoItem
//		(3) The database file will not be modified
func (t *ToDo) GetItem(id int) (ToDoItem, error) {

	// Check if item exists before trying to get it
	// this is a good practice, return an error if the
	// item does not exist
	item, ok := t.toDoMap[id]
	if !ok {
		return ToDoItem{}, errors.New("item does not exist")
	}

	return item, nil
}

// ChangeItemDoneStatus accepts an item id and a boolean status.
// It returns an error if the status could not be updated for any
// reason.  For example, the item itself does not exist, or an
// IO error trying to save the updated status.

// Preconditions:   (1) The database file must exist and be a valid
//
//					(2) The item must exist in the DB
//	    				because we use the item.Id as the key, this
//						function must check if the item already
//	    				exists in the DB, if not, return an error
//
// Postconditions:
//
//	 (1) The items status in the database will be updated
//		(2) If there is an error, it will be returned.
//		(3) This function MUST use existing functionality for most of its
//			work.  For example, it should call GetItem() to get the item
//			from the DB, then it should call UpdateItem() to update the
//			item in the DB (after the status is changed).
func (t *ToDo) ChangeItemDoneStatus(id int, value bool) error {

	//update was successful
	return errors.New("not implemented")
}

// GetAllItems returns all items from the DB.  If successful it
// returns a slice of all of the items to the caller
// Preconditions:   (1) The database file must exist and be a valid
//
// Postconditions:
//
//	 (1) All items will be returned, if any exist
//		(2) If there is an error, it will be returned
//			along with an empty slice
//		(3) The database file will not be modified
func (t *ToDo) GetAllItems() ([]ToDoItem, error) {

	//Now that we have the DB loaded, lets crate a slice
	var toDoList []ToDoItem

	//Now lets iterate over our map and add each item to our slice
	for _, item := range t.toDoMap {
		toDoList = append(toDoList, item)
	}

	//Now that we have all of our items in a slice, return it
	return toDoList, nil
}

// PrintItem accepts a ToDoItem and prints it to the console
// in a JSON pretty format. As some help, look at the
// json.MarshalIndent() function from our in class go tutorial.
func (t *ToDo) PrintItem(item ToDoItem) {
	jsonBytes, _ := json.MarshalIndent(item, "", "  ")
	fmt.Println(string(jsonBytes))
}

// PrintAllItems accepts a slice of ToDoItems and prints them to the console
// in a JSON pretty format.  It should call PrintItem() to print each item
// versus repeating the code.
func (t *ToDo) PrintAllItems(itemList []ToDoItem) {
	for _, item := range itemList {
		t.PrintItem(item)
	}
}

// JsonToItem accepts a json string and returns a ToDoItem
// This is helpful because the CLI accepts todo items for insertion
// and updates in JSON format.  We need to convert it to a ToDoItem
// struct to perform any operations on it.
func (t *ToDo) JsonToItem(jsonString string) (ToDoItem, error) {
	var item ToDoItem
	err := json.Unmarshal([]byte(jsonString), &item)
	if err != nil {
		return ToDoItem{}, err
	}

	return item, nil
}

*/