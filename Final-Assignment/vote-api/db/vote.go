package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
)

const (
	RedisNilError = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisVoteKeyPrefix = "vote:"
	VotersDefaultLocation = "0.0.0.0:1081"
	PollsDefaultLocation = "0.0.0.0:1082"
)

type cache struct {
	cacheClient *redis.Client
	jsonHelper *rejson.Handler
	context context.Context
}

type Vote struct {
	VoteID uint
	Voter string
	Poll string
	PollOption string
	VoteDate time.Time
}

type VoteDetails struct {
	VoteID uint
	Voter Voter
	Poll Poll
	PollOption PollOption
	VoteDate time.Time
}

type VoteKeys struct {
	VoteID uint
	VoterID uint
	PollID uint
	PollOptionID uint
}

type Voter struct {
	VoterID uint
	FirstName string
	LastName string
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


type VoteData struct {
	cache
	votersUrl string
	pollsUrl string
}

// Creat New Vote Data Handler 
func New() (*VoteData, error){

	redisUrl := os.Getenv("REDIS_URL")

	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}

	return NewWithCacheInstance(redisUrl)
}

func NewWithCacheInstance(location string) (*VoteData, error) {
	
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

	return &VoteData{
		cache: cache{
			cacheClient: client,
			jsonHelper: jsonHelper,
			context: ctx,
		},
		votersUrl: getVotersUrl(),
		pollsUrl: getPollsUrl(),
	}, nil
}

func getVotersUrl() string {
	votersUrl := os.Getenv("VOTERS_URL")

	if votersUrl == "" {
		votersUrl = VotersDefaultLocation
	}

	return votersUrl
}

func getPollsUrl() string {
	pollsUrl := os.Getenv("POLLS_URL")

	if pollsUrl == ""{
		pollsUrl = PollsDefaultLocation
	}

	return pollsUrl
}

func isRedisNilError(err error) bool {
	return errors.Is(err, redis.Nil) || err.Error() == RedisNilError
}

func redisVoteKeyFromId(id int) string {
	return fmt.Sprintf("%s%d", RedisVoteKeyPrefix, id)
}

func (v *VoteData) getVoteFromRedis(key string, vote *Vote) error {
	voteObject, err := v.jsonHelper.JSONGet(key, ".")
	if err != nil {
		return err
	}

	err = json.Unmarshal(voteObject.([]byte), vote)
	if err != nil {
		return err
	}

	return nil
}

func (v *VoteData) NewVote(voteID uint, voterID uint, pollID uint, pollOptionID uint) (*Vote, error){
	voter := &Vote{
		VoteID: voteID,
		Voter: v.getVoterUrl(voterID),
		Poll: v.getPollUrl(pollID),
		PollOption: v.getPollOptionUrl(pollID, pollOptionID),
		VoteDate: time.Now(),
	}

	return voter, nil
}

func (v *VoteData) getVoterUrl(voterID uint) string {
	return "http://" + v.votersUrl + "/voters/" + strconv.FormatUint(uint64(voterID), 10)
}

func (v *VoteData) getPollUrl(pollID uint) string {
	return "http://" + v.pollsUrl + "/polls/" + strconv.FormatUint(uint64(pollID), 10)
}

func (v *VoteData) getPollOptionUrl(pollID uint, optionID uint) string {
	url :=  "http://" + v.pollsUrl + "/polls/" + strconv.FormatUint(uint64(pollID), 10) 
	url += "/polloption/" + strconv.FormatUint(uint64(optionID), 10) 
	return url
}

func (v *VoteData) GetAllVotes() ([]Vote, error){
	var voters []Vote
	var voter Vote

	pattern := RedisVoteKeyPrefix + "*"
	ks, _ := v.cacheClient.Keys(v.context, pattern).Result()
	for _,key := range ks {
		err := v.getVoteFromRedis(key, &voter)
		if err != nil {
			return nil, err
		}
		voters = append(voters, voter)
	}
	return voters, nil
} 

func (v *VoteData) GetVote(voterID uint) (Vote, error){
	
	var vote Vote
	pattern := redisVoteKeyFromId(int(voterID))
	err := v.getVoteFromRedis(pattern, &vote)
	if err != nil {
		return Vote{}, err
	}

	return vote, nil
} 

func (v *VoteData) GetVoteDetails(voteID uint) (VoteDetails, error){
	
	var vote Vote
	pattern := redisVoteKeyFromId(int(voteID))
	err := v.getVoteFromRedis(pattern, &vote)
	if err != nil {
		return VoteDetails{}, err
	}
	
	voterDetails, err := getVoterDetails(vote)
	if err != nil { return VoteDetails{}, err}

	pollDetails, err := getPollDetails(vote)
	if err != nil { return VoteDetails{}, err}

	pollOptionDetails, err := getPollOptionDetails(vote)
	if err != nil { return VoteDetails{}, err}

	voteDetails := VoteDetails{
		VoteID: vote.VoteID,
		Voter: voterDetails,
		Poll: pollDetails,
		PollOption: pollOptionDetails,
		VoteDate: vote.VoteDate,
	}
	return voteDetails, nil
} 

func getVoterDetails(vote Vote) (Voter, error){
	resp, err := http.Get(vote.Voter)
	if err != nil {
		return Voter{}, errors.New("Error: failed request from voter service: " + vote.Voter + err.Error())
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Voter{}, errors.New("Error: could not read body from voter request")
	}

	var voter Voter
	err = json.Unmarshal([]byte(body), &voter)
	if err != nil {
		return Voter{}, err
	}

	return voter, nil
}

func getPollDetails(vote Vote) (Poll, error){
	resp, err := http.Get(vote.Poll)
	if err != nil {
		return Poll{}, errors.New("Error: could not get poll details")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Poll{}, errors.New("Error: could not get poll details")
	}

	var poll Poll
	err = json.Unmarshal([]byte(body), &poll)
	if err != nil {
		return Poll{}, err
	}
	return poll, nil
}

func getPollOptionDetails(vote Vote) (PollOption, error){
	resp, err := http.Get(vote.PollOption)
	if err != nil {
		return PollOption{}, errors.New("Error: could not get poll option details")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return PollOption{}, errors.New("Error: could not get poll option details")
	}
	var pollOption PollOption
	err = json.Unmarshal([]byte(body), &pollOption)
	if err != nil {
		return PollOption{}, err
	}
	return pollOption, nil
}

func (v *VoteData) AddVote(voteKeys VoteKeys) error {

	redisKey := redisVoteKeyFromId(int(voteKeys.VoteID))
	var existingItem Vote
	if err := v.getVoteFromRedis(redisKey, &existingItem); err == nil {
		return errors.New("item already exists")
	}

	newVote, _ := v.NewVote(voteKeys.VoteID, voteKeys.VoterID, voteKeys.PollID, voteKeys.PollOptionID)

	if _, err := v.jsonHelper.JSONSet(redisKey, ".", newVote); err != nil {
		return err
	}

	return nil
}

func (v *VoteData) UpdateVote(voteID uint, updateData VoteKeys) error {

	redisKey := redisVoteKeyFromId(int(voteID))
	var existingVote Vote
	if err := v.getVoteFromRedis(redisKey, &existingVote); err != nil {
		return errors.New("Item does not exist")
	}

	updatedVote,_ := v.NewVote(updateData.VoteID, 
		updateData.VoterID, 
		updateData.PollID, 
		updateData.PollOptionID)

	if _, err := v.jsonHelper.JSONSet(redisKey, ".", updatedVote); err != nil {
		return err
	}

	return nil
}

func (v *VoteData) DeleteVote(voteID uint) error {
	pattern := redisVoteKeyFromId(int(voteID))
	numDeleted, err := v.cacheClient.Del(v.context, pattern).Result()
	if err != nil {
		return err
	}
	if numDeleted == 0 {
		return errors.New("Attempted to delete a non-existent vote")
	}

	return nil
}



