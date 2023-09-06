package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
	"log"
	"os"
)

const (
	RedisNilError        = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisKeyPrefix       = "voter:"
	RedisIDKey           = "voterCnt:"
)

type Voter struct {
	VoterID     uint   `json:"id"`
	FirstName   string `json:"FirstName"`
	LastName    string `json:"LastName"`
	VoteHistory []uint `json:"pollid"`
}

type VoterAPI struct {
	cacheClient *redis.Client
	jsonHelper  *rejson.Handler
	context     context.Context
	idCnter     uint
}

func NewVoterApi() (*VoterAPI, error) {
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}

	api, err := NewDbWithInstance(redisUrl)
	if err != nil {
		log.Println("Failed to initialize the voter API!")
		return &VoterAPI{}, err
	}

	itemObject, err := api.jsonHelper.JSONGet(RedisIDKey, ".")
	if err != nil {
		// There's no entry for the current number of voters,
		// assume 0
		return api, nil
	}

	//JSONGet returns an "any" object, or empty interface,
	//we need to convert it to a byte array, which is the
	//underlying type of the object, then we can unmarshal
	//it into our ToDoItem struct
	err = json.Unmarshal(itemObject.([]byte), &api.idCnter)
	if err != nil {
		return &VoterAPI{}, err
	}

	return api, nil
}

func NewDbWithInstance(location string) (*VoterAPI, error) {

	client := redis.NewClient(&redis.Options{
		Addr: location,
	})

	ctx := context.Background()

	err := client.Ping(ctx).Err()
	if err != nil {
		log.Println("Error connecting to redis" + err.Error())
		return nil, err
	}

	jsonHelper := rejson.NewReJSONHandler()
	jsonHelper.SetGoRedisClientWithContext(ctx, client)

	return &VoterAPI{
			cacheClient: client,
			jsonHelper:  jsonHelper,
			context:     ctx,
		},
		nil
}

//------------------------------------------------------------
// REDIS HELPERS
//------------------------------------------------------------

// In redis, our keys will be strings, they will look like
// todo:<number>.  This function will take an integer and
// return a string that can be used as a key in redis
func redisKeyFromId(id int) string {
	return fmt.Sprintf("%s%d", RedisKeyPrefix, id)
}

func (t *VoterAPI) getVoterFromRedis(key string, item *Voter) error {

	//Lets query redis for the item, note we can return parts of the
	//json structure, the second parameter "." means return the entire
	//json structure
	itemObject, err := t.jsonHelper.JSONGet(key, ".")
	if err != nil {
		return err
	}

	//JSONGet returns an "any" object, or empty interface,
	//we need to convert it to a byte array, which is the
	//underlying type of the object, then we can unmarshal
	//it into our ToDoItem struct
	err = json.Unmarshal(itemObject.([]byte), item)
	if err != nil {
		return err
	}

	return nil
}

func (t *VoterAPI) AddVoter(fn string, ln string) (*Voter, error) {

	//Before we add an item to the DB, lets make sure
	//it does not exist, if it does, return an error
	redisKey := redisKeyFromId(int(t.idCnter + 1))
	var existingItem Voter
	if err := t.getVoterFromRedis(redisKey, &existingItem); err == nil {
		return &Voter{}, errors.New("Voter already exists!")
	}

	newVoter := Voter{
		VoterID:     t.idCnter + 1,
		FirstName:   fn,
		LastName:    ln,
		VoteHistory: []uint{},
	}

	//Add item to database with JSON Set
	if _, err := t.jsonHelper.JSONSet(redisKey, ".", newVoter); err != nil {
		return &Voter{}, err
	}

	//Increment the API counter only after we have succesfully added a new voter to the DB
	t.idCnter += 1

	if _, err := t.jsonHelper.JSONSet(RedisIDKey, ".", t.idCnter); err != nil {
		return &Voter{}, err
	}

	//If everything is ok, return nil for the error
	return &newVoter, nil
}

func (t *VoterAPI) GetVoter(id int) (Voter, error) {

	// Check if item exists before trying to get it
	// this is a good practice, return an error if the
	// item does not exist
	var voter Voter
	pattern := redisKeyFromId(id)
	err := t.getVoterFromRedis(pattern, &voter)
	if err != nil {
		return Voter{}, err
	}

	return voter, nil
}

func (t *VoterAPI) GetAllVoters() ([]Voter, error) {

	//Now that we have the DB loaded, lets crate a slice
	var voterList []Voter
	var vtr Voter

	//Lets query redis for all of the items
	pattern := RedisKeyPrefix + "*"
	ks, _ := t.cacheClient.Keys(t.context, pattern).Result()
	for _, key := range ks {
		err := t.getVoterFromRedis(key, &vtr)
		if err != nil {
			return nil, err
		}
		voterList = append(voterList, vtr)
	}

	return voterList, nil
}
