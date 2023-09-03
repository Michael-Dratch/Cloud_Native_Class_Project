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
	RedisVoterKeyPrefix = "voter:"
)

type cache struct {
	cacheClient *redis.Client
	jsonHelper *rejson.Handler
	context context.Context
}

type Voter struct {
	VoterID uint
	FirstName string
	LastName string
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
	}

	return voter, nil
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