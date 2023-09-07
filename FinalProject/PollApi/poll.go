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
	RedisKeyPrefix       = "poll:"
	RedisIDKey           = "pollCnt:"
)

type Poll struct {
	PollID       uint     `json:"pollID"`
	PollTitle    string   `json:"pollTitle"`
	PollQuestion string   `json:"pollQuestion"`
	PollOptions  []string `json:"pollOptions"`
}

type PollApi struct {
	cacheClient *redis.Client
	jsonHelper  *rejson.Handler
	context     context.Context
	idCnter     uint
}

func NewPollApi() (*PollApi, error) {
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}

	api, err := NewDbWithInstance(redisUrl)
	if err != nil {
		return &PollApi{}, err
	}

	itemObject, err := api.jsonHelper.JSONGet(RedisIDKey, ".")
	if err != nil {
		// There's no entry for the current number of polls,
		// assume 0
		return api, nil
	}

	//JSONGet returns an "any" object, or empty interface,
	//we need to convert it to a byte array, which is the
	//underlying type of the object, then we can unmarshal
	//it into our ToDoItem struct
	err = json.Unmarshal(itemObject.([]byte), &api.idCnter)
	if err != nil {
		return &PollApi{}, err
	}

	return api, nil
}

func NewDbWithInstance(location string) (*PollApi, error) {

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
	return &PollApi{
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

func (t *PollApi) getPollFromRedis(key string, poll *Poll) error {

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
	err = json.Unmarshal(itemObject.([]byte), poll)
	if err != nil {
		return err
	}

	return nil
}

func (t *PollApi) AddPoll(pollTitle string, pollQuestion string, pollOptions []string) (*Poll, error) {

	//Before we add an item to the DB, lets make sure
	//it does not exist, if it does, return an error
	redisKey := redisKeyFromId(int(t.idCnter + 1))
	var existingPoll Poll
	if err := t.getPollFromRedis(redisKey, &existingPoll); err == nil {
		return &Poll{}, errors.New("Poll already exists!")
	}

	newPoll := Poll{
		PollID:       t.idCnter + 1,
		PollTitle:    pollTitle,
		PollQuestion: pollQuestion,
		PollOptions:  pollOptions,
	}

	//Add item to database with JSON Set
	if _, err := t.jsonHelper.JSONSet(redisKey, ".", newPoll); err != nil {
		return &Poll{}, err
	}

	t.idCnter += 1

	if _, err := t.jsonHelper.JSONSet(RedisIDKey, ".", t.idCnter); err != nil {
		return &Poll{}, err
	}

	//If everything is ok, return nil for the error
	return &newPoll, nil
}

func (t *PollApi) GetPoll(pollID int) (*Poll, error) {

	// Check if item exists before trying to get it
	// this is a good practice, return an error if the
	// item does not exist
	var poll Poll
	pattern := redisKeyFromId(pollID)
	err := t.getPollFromRedis(pattern, &poll)
	if err != nil {
		return &Poll{}, err
	}

	return &poll, nil
}

func (t *PollApi) GetAllPolls() ([]Poll, error) {

	//Now that we have the DB loaded, lets crate a slice
	var pollList []Poll
	var poll Poll

	//Lets query redis for all of the items
	pattern := RedisKeyPrefix + "*"
	ks, _ := t.cacheClient.Keys(t.context, pattern).Result()
	for _, key := range ks {
		err := t.getPollFromRedis(key, &poll)
		if err != nil {
			return nil, err
		}
		pollList = append(pollList, poll)
	}

	return pollList, nil
}
