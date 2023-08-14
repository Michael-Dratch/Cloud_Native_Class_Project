package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
)

const (
	RedisNilError = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisVoterKeyPrefix = "voter:"
	RedisVoterPollKeyPrefix = "voterPoll:"
)

type cache struct {
	cacheClient *redis.Client
	jsonHelper *rejson.Handler
	context context.Context
}

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

type VoterData struct {
	cache
}

// Creat New Voter Data Handler 
func New() (*VoterData, error){

	redisUrl := os.Getenv("REDIS_URL")

	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}

	return NewWithCacheInstance(redisUrl)

}

func NewWithCacheInstance(location string) (*VoterData, error) {
	
	client := redis.NewClient(&redis.Options{
		Addr:location,
	})

	ctx := context.Background()

	err := client.Ping(ctx).Err()
	if err != nil {
		log.Println("Error connecting to redis" + err.Error())
		return nil, err
	}

	jsonHelper := rejson.NewReJSONHandler()
	jsonHelper.SetGoRedisClientWithContext(ctx, client)

	return &VoterData{
		cache: cache{
			cacheClient: client,
			jsonHelper: jsonHelper,
			context: ctx,
		},
	}, nil
}

func isRedisNilError(err error) bool {
	return errors.Is(err, redis.Nil) || err.Error() == RedisNilError
}

func redisVoterKeyFromId(id int) string {
	return fmt.Sprintf("%s%d", RedisVoterKeyPrefix, id)
}

func redisVoterPollKeyFromId(id int) string {
	return fmt.Sprintf("%s%d", RedisVoterPollKeyPrefix, id)
}

func (v *VoterData) getVoterFromRedis(key string, voter *Voter) error {
	voterObject, err := v.jsonHelper.JSONGet(key, ".")
	if err != nil {
		return err
	}

	err = json.Unmarshal(voterObject.([]byte), voter)
	if err != nil {
		return err
	}

	return nil
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


func (v *VoterData) GetAllVoters() ([]Voter, error){
	var voters []Voter
	var voter Voter

	pattern := RedisVoterKeyPrefix + "*"
	ks, _ := v.cacheClient.Keys(v.context, pattern).Result()
	for _,key := range ks {
		err := v.getVoterFromRedis(key, &voter)
		if err != nil {
			return nil, err
		}
		voters = append(voters, voter)
	}
	return voters, nil
} 

func (v *VoterData) GetVoter(voterID uint) (Voter, error){
	
	var voter Voter
	pattern := redisVoterKeyFromId(int(voterID))
	err := v.getVoterFromRedis(pattern, &voter)
	if err != nil {
		return Voter{}, err
	}

	return voter, nil
} 

func (v *VoterData) AddVoter(voter Voter) error {

	redisKey := redisVoterKeyFromId(int(voter.VoterID))
	var existingItem Voter
	if err := v.getVoterFromRedis(redisKey, &existingItem); err == nil {
		return errors.New("item already exists")
	}

	newVoter, _ := NewVoter(voter.VoterID, voter.FirstName, voter.LastName)

	if _, err := v.jsonHelper.JSONSet(redisKey, ".", newVoter); err != nil {
		return err
	}

	return nil
}

func (v *VoterData) UpdateVoter(voterID uint, updateData Voter) error {

	redisKey := redisVoterKeyFromId(int(voterID))
	var existingVoter Voter
	if err := v.getVoterFromRedis(redisKey, &existingVoter); err != nil {
		return errors.New("Item does not exist")
	}

	updatedVoter:= removeZeroValuesFromUpdateData(existingVoter, updateData)

	if _, err := v.jsonHelper.JSONSet(redisKey, ".", updatedVoter); err != nil {
		return err
	}

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

func (v *VoterData) DeleteVoter(voterID uint) error {
	pattern := redisVoterKeyFromId(int(voterID))
	numDeleted, err := v.cacheClient.Del(v.context, pattern).Result()
	if err != nil {
		return err
	}
	if numDeleted == 0 {
		return errors.New("Attempted to delete a non-existent voter")
	}

	return nil
}

func (v *VoterData) GetVoterHistory(voterID uint) ([]VoterPoll, error){
	voter, err := v.GetVoter(voterID)
	if err != nil {
		return make([]VoterPoll, 0) , errors.New("Voter ID does not exist")
	}
	
	return voter.VoteHistory, nil
}

func (v *VoterData) GetVoterPoll(voterID uint, pollID uint) (VoterPoll, error){
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

func (v *VoterData) DoesVoterPollExist(voterID uint, pollID uint) bool {
	_, err := v.GetVoterPoll(voterID, pollID)
	if err == nil { return true } 
	return false 
}

func (v *VoterData) AddVoterPoll(voterID uint, newPoll VoterPoll) error{
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
	v.UpdateVoter(voterID, voter)
	return nil
}

func (v *VoterData) UpdateVoterPoll(voterID uint, pollID uint, updateData VoterPoll) error{
	
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
	v.UpdateVoter(voterID, voter)
	return nil
}

func removeZeroValuesFromPollUpdateData(oldData VoterPoll, updateData VoterPoll) VoterPoll{
	if updateData.VoteDate.IsZero() {
		updateData.VoteDate = oldData.VoteDate
	}
	return updateData
}

func (v *VoterData) DeleteVoterPoll(voterID uint, pollID uint) error{
	voter, err := v.GetVoter(voterID)
	if err != nil {
		return errors.New("Voter does not exists")
	}

	for index,poll := range voter.VoteHistory{
		if poll.PollID == pollID{
			voter.VoteHistory = append(voter.VoteHistory[:index], voter.VoteHistory[index+1:]... )
			break
		}
	}

	v.UpdateVoter(voterID, voter)
	return nil
}

