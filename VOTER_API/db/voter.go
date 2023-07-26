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

func (v *VoterList) UpdateVoter(voterID uint, updateData Voter) error {
	oldData, ok := v.Voters[voterID]
	if ok == false {
		return errors.New("Voter does not exists")
	}
	
	updatedVoter:= removeZeroValuesFromUpdateData(oldData, updateData)
	v.Voters[voterID] = updatedVoter

	return nil
}

func removeZeroValuesFromUpdateData(oldData Voter, updateData Voter) Voter {
	if updateData.FirstName == "" {
		updateData.FirstName = oldData.FirstName
	}
	if updateData.LastName == "" {
		updateData.LastName = oldData.LastName
	}
	if updateData.VoteHistory == nil {
		updateData.VoteHistory = oldData.VoteHistory
	}

	return updateData
}

func (v *VoterList) DeleteVoter(voterID uint) error {
	_, ok := v.Voters[voterID]
	if ok == false {
		return errors.New("Voter does not exists")
	}

	delete(v.Voters, voterID)
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

func (v *VoterList) UpdateVoterPoll(voterID uint, pollID uint, updateData VoterPoll) error{
	
	oldData,err := v.GetVoterPoll(voterID, pollID)
	if err != nil {
		return errors.New("Error: Voter poll does not exist")
	}

	updateData = removeZeroValuesFromPollUpdateData(oldData, updateData)
	voter,_ := v.GetVoter(voterID)
	for index,poll := range voter.VoteHistory{
		if poll.PollID == pollID{
			voter.VoteHistory[index] = updateData
			break
		}
	}

	v.Voters[voterID] = voter
	return nil
}

func removeZeroValuesFromPollUpdateData(oldData VoterPoll, updateData VoterPoll) VoterPoll{
	if updateData.VoteDate.IsZero() {
		updateData.VoteDate = oldData.VoteDate
	}
	return updateData
}

func (v *VoterList) DeleteVoterPoll(voterID uint, pollID uint) error{
	voter, ok := v.Voters[voterID]
	if ok == false {
		return errors.New("Voter does not exists")
	}

	for index,poll := range voter.VoteHistory{
		if poll.PollID == pollID{
			voter.VoteHistory = append(voter.VoteHistory[:index], voter.VoteHistory[index+1:]... )
			break
		}
	}

	v.Voters[voterID] = voter
	return nil
}

