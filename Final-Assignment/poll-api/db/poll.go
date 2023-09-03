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
	PollOptions []PollOption
}

type PollOption struct {
	PollOptionID uint
	PollOptionText string
}



type PollData struct {
	cache
}

// Creat New Voter Data Handler 
func New() (*PollData, error){

	redisUrl := os.Getenv("REDIS_URL")

	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}

	return NewWithCacheInstance(redisUrl)

}

func NewWithCacheInstance(location string) (*PollData, error) {
	
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

	return &PollData{
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

func (p *PollData) getPollFromRedis(key string, poll *Poll) error {
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
		PollOptions:make ([]PollOption, 0),
	}

	return poll, nil
}

func NewPollOption(	pollOptionID uint, pollOptionText string) (*PollOption, error){
	pollOption:= &PollOption{
		PollOptionID: pollOptionID,
		PollOptionText: pollOptionText,
	}
	return pollOption, nil
}


func (p *PollData) GetAllPolls() ([]Poll, error){
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

func (p *PollData) GetPoll(pollID uint) (Poll, error){
	
	var poll Poll
	pattern := redisPollKeyFromId(int(pollID))
	err := p.getPollFromRedis(pattern, &poll)
	if err != nil {
		return Poll{}, err
	}

	return poll, nil
} 

func (p *PollData) AddPoll(poll Poll) error {

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

func (p *PollData) UpdatePoll(pollID uint, updateData Poll) error {

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

func (p *PollData) DeletePoll(pollID uint) error {
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

func (p *PollData) GetPollOptions(pollID uint) ([]PollOption, error){
	poll, err := p.GetPoll(pollID)
	if err != nil {
		return make([]PollOption, 0) , errors.New("Poll ID does not exist")
	}
	
	return poll.PollOptions, nil
}

func (p *PollData) GetPollOption(pollID uint, pollOptionID uint) (PollOption, error){
	poll, err := p.GetPoll(pollID)
	if err != nil {
		return PollOption{} , errors.New("Poll ID does not exist")
	} 

	for _, pollOption := range poll.PollOptions{
		if pollOption.PollOptionID == pollOptionID{
			return pollOption, nil
		}
	}
	log.Println("Error: Poll option ID does not exist for this poll")
	return PollOption{}, errors.New("Poll option ID does not exist for this poll")
}

func (p *PollData) DoesPollOptionExist(pollID uint, pollOptionID uint) bool {
	_, err := p.GetPollOption(pollID, pollOptionID)
	if err == nil { return true } 
	return false 
}

func (p *PollData) AddPollOption(pollID uint, newPollOption PollOption) error{
	poll, err := p.GetPoll(pollID)
	if err != nil {
		return errors.New("Poll ID does not exist")
	} 

	for _, pollOption := range poll.PollOptions{
		if newPollOption.PollOptionID == pollOption.PollOptionID{
			return errors.New("Poll Option ID already exists for this poll")
		}
	}

	poll.PollOptions = append(poll.PollOptions, newPollOption)
	p.UpdatePoll(pollID, poll)
	return nil
}

func (p *PollData) UpdatePollOption(pollID uint, pollOptionID uint, updateData PollOption) error{
	
	oldData,err := p.GetPollOption(pollID, pollOptionID)
	if err != nil {
		return errors.New("Error: Poll option does not exist")
	}

	updateData = removeZeroValuesFromPollOptionUpdateData(oldData, updateData)
	poll,_ := p.GetPoll(pollID)
	for index, pollOption := range poll.PollOptions{
		if pollOption.PollOptionID == pollOptionID{
			poll.PollOptions[index] = updateData
			break
		}
	}
	p.UpdatePoll(pollID, poll)
	return nil
}

func removeZeroValuesFromPollOptionUpdateData(oldData PollOption, updateData PollOption) PollOption{
	if updateData.PollOptionText == "" {
		updateData.PollOptionText = oldData.PollOptionText
	}
	return updateData
}

func (p *PollData) DeletePollOption(pollID uint, pollOptionID uint) error{
	poll, err := p.GetPoll(pollID)
	if err != nil {
		return errors.New("Poll does not exists")
	}

	for index,pollOption := range poll.PollOptions{
		if pollOption.PollOptionID == pollOptionID{
			poll.PollOptions = append(poll.PollOptions[:index], poll.PollOptions[index+1:]... )
			break
		}
	}

	p.UpdatePoll(pollID, poll)
	return nil
}

