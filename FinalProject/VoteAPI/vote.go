package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/go-resty/resty/v2"
	"github.com/nitishm/go-rejson/v4"
	"log"
	"os"
)

const (
	RedisNilError        = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisKeyPrefix       = "vote:"
	RedisIDKey           = "voteCnt:"
	VoterDefaultLocation = "0.0.0.0:2080"
	PollDefaultLocation  = "0.0.0.0:3080"
)

type Vote struct {
	VoteID    uint `json:"voteID"`
	VoterID   uint `json:"voterID"`
	PollID    uint `json:"pollID"`
	VoteValue uint `json:"voteValue"`
}

type VoteApi struct {
	cacheClient *redis.Client
	jsonHelper  *rejson.Handler
	context     context.Context
	apiClient   *resty.Client
	VoterUrl    string
	PollUrl     string
	idCnter     uint
}

func NewVoteApi() (*VoteApi, error) {
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}

	voterUrl := os.Getenv("VOTER_URL")
	if redisUrl == "" {
		redisUrl = VoterDefaultLocation
	}

	pollUrl := os.Getenv("POLL_URL")
	if pollUrl == "" {
		pollUrl = PollDefaultLocation
	}

	api, err := NewDbWithInstance(redisUrl)
	if err != nil {
		return &VoteApi{}, err
	}

	api.apiClient = resty.New()
	api.VoterUrl = voterUrl
	api.PollUrl = pollUrl

	itemObject, err := api.jsonHelper.JSONGet(RedisIDKey, ".")
	if err != nil {
		// There's no entry for the current number of voter,
		// assume 0
		return api, nil
	}

	//JSONGet returns an "any" object, or empty interface,
	//we need to convert it to a byte array, which is the
	//underlying type of the object, then we can unmarshal
	//it into our ToDoItem struct
	err = json.Unmarshal(itemObject.([]byte), &api.idCnter)
	if err != nil {
		return &VoteApi{}, err
	}

	return api, nil
}

func NewDbWithInstance(location string) (*VoteApi, error) {

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

	//Return a pointer to a new ToDo struct
	return &VoteApi{
			cacheClient: client,
			jsonHelper:  jsonHelper,
			context:     ctx,
		},
		nil
}

// In redis, our keys will be strings, they will look like
// todo:<number>.  This function will take an integer and
// return a string that can be used as a key in redis
func redisKeyFromId(id int) string {
	return fmt.Sprintf("%s%d", RedisKeyPrefix, id)
}

func (t *VoteApi) getVoteFromRedis(key string, vote *Vote) error {

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
	err = json.Unmarshal(itemObject.([]byte), vote)
	if err != nil {
		return err
	}

	return nil
}

func (t *VoteApi) AddVote(voterID uint, pollID uint, value uint) (*Vote, error) {

	//Before we add an item to the DB, lets make sure
	//it does not exist, if it does, return an error
	redisKey := redisKeyFromId(int(t.idCnter + 1))
	var existingItem Vote
	if err := t.getVoteFromRedis(redisKey, &existingItem); err == nil {
		return &Vote{}, errors.New("Vote already exists!")
	}

	// Make sure that the voter exists
	voterUrl := fmt.Sprint(t.VoterUrl, "/voter/", voterID)
	resp, err := t.apiClient.R().Get(voterUrl)
	if err != nil {
		log.Println("Error when trying to reach voter api: ", voterUrl)
		return &Vote{}, err
	}

	if resp.StatusCode() == 404 {
		return &Vote{}, errors.New("The voter submitting a vote does not exist!!")
	}

	// Make sure that the poll exists
	pollUrl := fmt.Sprint(t.PollUrl, "/poll/", pollID)
	resp, err = t.apiClient.R().Get(pollUrl)
	if err != nil {
		log.Println("Error when trying to reach poll api: ", pollUrl)
		return &Vote{}, err
	}

	if resp.StatusCode() == 404 {
		return &Vote{}, errors.New("The poll you are trying to vote in does not exist!!")
	}

	voterNewPollUrl := fmt.Sprintf(voterUrl, "/", pollID)
	resp, err = t.apiClient.R().SetHeader("Content-Type", "application/json").Post(voterNewPollUrl)
	if err != nil {
		return &Vote{}, errors.New("Could not connect to voter-api to post new poll history")
	}

	newVote := Vote{
		VoteID:    t.idCnter + 1,
		VoterID:   voterID,
		PollID:    pollID,
		VoteValue: value,
	}

	//Add item to database with JSON Set
	if _, err := t.jsonHelper.JSONSet(redisKey, ".", newVote); err != nil {
		return &Vote{}, err
	}

	t.idCnter += 1

	if _, err := t.jsonHelper.JSONSet(RedisIDKey, ".", t.idCnter); err != nil {
		return &Vote{}, err
	}

	//If everything is ok, return nil for the error
	return &newVote, nil
}

func (t *VoteApi) GetVote(voteID int) (Vote, error) {

	// Check if item exists before trying to get it
	// this is a good practice, return an error if the
	// item does not exist
	var vote Vote
	pattern := redisKeyFromId(voteID)
	err := t.getVoteFromRedis(pattern, &vote)
	if err != nil {
		return Vote{}, err
	}

	return vote, nil
}

func (t *VoteApi) GetAllVotes() ([]Vote, error) {

	//Now that we have the DB loaded, lets crate a slice
	var voteList []Vote
	var vt Vote

	//Lets query redis for all of the items
	pattern := RedisKeyPrefix + "*"
	ks, _ := t.cacheClient.Keys(t.context, pattern).Result()
	for _, key := range ks {
		err := t.getVoteFromRedis(key, &vt)
		if err != nil {
			return nil, err
		}
		voteList = append(voteList, vt)
	}

	return voteList, nil
}
