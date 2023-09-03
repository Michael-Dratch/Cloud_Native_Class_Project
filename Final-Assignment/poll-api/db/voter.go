package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
)

const (
	RedisNilError = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisPollKeyPrefix = "poll:"
)

type cache struct {
	cacheClient *redis.Client
	jsonHelper *rejson.Handler
	context context.Context
}

type Poll struct {
	PollID uint
	PollTitle string
	PollQuestion string
	PollOptions []pollOption
}

type pollOption struct {
	PollOptionID uint
	PollOptionText string
}



type PollsData struct {
	cache
}

// Creat New Voter Data Handler 
func New() (*PollsData, error){

	redisUrl := os.Getenv("REDIS_URL")

	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}

	return NewWithCacheInstance(redisUrl)

}

func NewWithCacheInstance(location string) (*PollsData, error) {
	
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

	return &PollsData{
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


func redisPollKeyFromId(id int) string {
	return fmt.Sprintf("%s%d", RedisPollKeyPrefix, id)
}

func (p *PollsData) getPollFromRedis(key string, poll *Poll) error {
	pollObject, err := p.jsonHelper.JSONGet(key, ".")
	if err != nil {
		return err
	}

	err = json.Unmarshal(pollObject.([]byte), poll)
	if err != nil {
		return err
	}

	return nil
}

func NewPoll(pollID uint, pollTitle string, pollQuestion string) (*Poll, error){
	poll := &Poll{
		PollID: pollID,
		PollTitle: pollTitle,
		PollQuestion: pollQuestion,
		PollOptions:make ([]pollOption, 0),
	}

	return poll, nil
}

func NewPollOption(	pollOptionID uint, pollOptionText string) (*pollOption, error){
	pollOption:= &pollOption{
		PollOptionID: pollOptionID,
		PollOptionText: pollOptionText,
	}
	return pollOption, nil
}


func (p *PollsData) GetAllPolls() ([]Poll, error){
	var polls []Poll
	var poll Poll

	pattern := RedisPollKeyPrefix + "*"
	ks, _ := p.cacheClient.Keys(p.context, pattern).Result()
	for _,key := range ks {
		err := p.getPollFromRedis(key, &poll)
		if err != nil {
			return nil, err
		}
		polls = append(polls, poll)
	}
	return polls, nil
} 

func (p *PollsData) GetPoll(pollID uint) (Poll, error){
	
	var poll Poll
	pattern := redisPollKeyFromId(int(pollID))
	err := p.getPollFromRedis(pattern, &poll)
	if err != nil {
		return Poll{}, err
	}

	return poll, nil
} 

func (p *PollsData) AddPoll(poll Poll) error {

	redisKey := redisPollKeyFromId(int(poll.PollID))
	var existingItem Poll
	if err := p.getPollFromRedis(redisKey, &existingItem); err == nil {
		return errors.New("item already exists")
	}

	newPoll, _ := NewPoll(poll.PollID, poll.PollTitle, poll.PollQuestion)

	if _, err := p.jsonHelper.JSONSet(redisKey, ".", newPoll); err != nil {
		return err
	}

	return nil
}

func (p *PollsData) UpdatePoll(pollID uint, updateData Poll) error {

	redisKey := redisPollKeyFromId(int(pollID))
	var existingPoll Poll
	if err := p.getPollFromRedis(redisKey, &existingPoll); err != nil {
		return errors.New("Item does not exist")
	}

	updatedPoll:= removeZeroValuesFromUpdateData(existingPoll, updateData)

	if _, err := p.jsonHelper.JSONSet(redisKey, ".", updatedPoll); err != nil {
		return err
	}

	return nil
}

func removeZeroValuesFromUpdateData(oldData Poll, updateData Poll) Poll {
	if updateData.PollTitle == "" {
		updateData.PollTitle = oldData.PollTitle
	}
	if updateData.PollQuestion == "" {
		updateData.PollQuestion = oldData.PollQuestion
	}
	if updateData.PollOptions == nil {
		updateData.PollOptions = oldData.PollOptions
	}

	return updateData
}

func (p *PollsData) DeletePoll(pollID uint) error {
	pattern := redisPollKeyFromId(int(pollID))
	numDeleted, err := p.cacheClient.Del(p.context, pattern).Result()
	if err != nil {
		return err
	}
	if numDeleted == 0 {
		return errors.New("Attempted to delete a non-existent poll")
	}

	return nil
}

func (p *PollsData) GetPollOptions(pollID uint) ([]pollOption, error){
	poll, err := p.GetPoll(pollID)
	if err != nil {
		return make([]pollOption, 0) , errors.New("Poll ID does not exist")
	}
	
	return poll.PollOptions, nil
}

func (p *PollsData) GetPollOption(pollID uint, pollOptionID uint) (pollOption, error){
	poll, err := p.GetPoll(pollID)
	if err != nil {
		return pollOption{} , errors.New("Poll ID does not exist")
	} 

	for _, pollOption := range poll.PollOptions{
		if pollOption.PollOptionID == pollOptionID{
			return pollOption, nil
		}
	}
	log.Println("Error: Poll option ID does not exist for this poll")
	return pollOption{}, errors.New("Poll option ID does not exist for this poll")
}

func (p *PollsData) DoesVoterPollExist(voterID uint, pollID uint) bool {
	_, err := p.GetPollOption(pollID, pollOptionID)
	if err == nil { return true } 
	return false 
}

func (p *PollsData) AddVoterPoll(voterID uint, newPoll VoterPoll) error{
	voter, err := p.GetVoter(voterID)
	if err != nil {
		return errors.New("Voter ID does not exist")
	} 

	for _, poll := range voter.VoteHistory{
		if newPoll.PollID == poll.PollID{
			return errors.New("Poll ID already exists for this voter")
		}
	}

	voter.VoteHistory = append(voter.VoteHistory, newPoll)
	p.UpdateVoter(voterID, voter)
	return nil
}

func (p *PollsData) UpdateVoterPoll(voterID uint, pollID uint, updateData VoterPoll) error{
	
	oldData,err := p.GetVoterPoll(voterID, pollID)
	if err != nil {
		return errors.New("Error: Voter poll does not exist")
	}

	updateData = removeZeroValuesFromPollUpdateData(oldData, updateData)
	voter,_ := p.GetVoter(voterID)
	for index,poll := range voter.VoteHistory{
		if poll.PollID == pollID{
			voter.VoteHistory[index] = updateData
			break
		}
	}
	p.UpdateVoter(voterID, voter)
	return nil
}

func removeZeroValuesFromPollUpdateData(oldData VoterPoll, updateData VoterPoll) VoterPoll{
	if updateData.VoteDate.IsZero() {
		updateData.VoteDate = oldData.VoteDate
	}
	return updateData
}

func (p *PollsData) DeleteVoterPoll(voterID uint, pollID uint) error{
	voter, err := p.GetVoter(voterID)
	if err != nil {
		return errors.New("Voter does not exists")
	}

	for index,poll := range voter.VoteHistory{
		if poll.PollID == pollID{
			voter.VoteHistory = append(voter.VoteHistory[:index], voter.VoteHistory[index+1:]... )
			break
		}
	}

	p.UpdateVoter(voterID, voter)
	return nil
}

