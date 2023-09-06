package vote

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
	RedisKeyPrefix       = "vote:"
)

type Vote struct {
	VoteID    uint
	VoterID   uint
	PollID    uint
	VoteValue uint
}

type VoteDb struct {
	cacheClient *redis.Client
	jsonHelper  *rejson.Handler
	context     context.Context
}

func NewVote(voteid uint, voterid uint, pollid uint, value uint) *Vote {
	return &Vote{
		VoteID:    voteid,
		VoterID:   voterid,
		PollID:    pollid,
		VoteValue: value,
	}
}

func NewDb() (*VoteDb, error) {
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}

	return NewDbWithInstance(redisUrl)
}

func NewDbWithInstance(location string) (*VoteDb, error) {

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
	return &VoteDb{
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

func (t *VoteDb) getVoteFromRedis(key string, vote *Vote) error {

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

func (t *VoteDb) AddVote(vote Vote) error {

	//Before we add an item to the DB, lets make sure
	//it does not exist, if it does, return an error
	redisKey := redisKeyFromId(int(vote.VoteID))
	var existingItem Vote
	if err := t.getVoteFromRedis(redisKey, &existingItem); err == nil {
		return errors.New("Vote already exists!")
	}

	//Add item to database with JSON Set
	if _, err := t.jsonHelper.JSONSet(redisKey, ".", vote); err != nil {
		return err
	}

	//If everything is ok, return nil for the error
	return nil
}

func (t *VoteDb) GetVote(id int) (Vote, error) {

	// Check if item exists before trying to get it
	// this is a good practice, return an error if the
	// item does not exist
	var vote Vote
	pattern := redisKeyFromId(id)
	err := t.getVoteFromRedis(pattern, &vote)
	if err != nil {
		return Vote{}, err
	}

	return vote, nil
}

func (t *VoteDb) UpdateVote(vote Vote) error {

	//Before we add an item to the DB, lets make sure
	//it does not exist, if it does, return an error
	redisKey := redisKeyFromId(int(vote.VoteID))
	var existingVote Vote
	if err := t.getVoteFromRedis(redisKey, &existingVote); err != nil {
		return errors.New("vote does not exist")
	}

	//Add item to database with JSON Set.  Note there is no update
	//functionality, so we just overwrite the existing item
	if _, err := t.jsonHelper.JSONSet(redisKey, ".", vote); err != nil {
		return err
	}

	//If everything is ok, return nil for the error
	return nil
}

func (t *VoteDb) GetAllVotes() ([]Vote, error) {

	//Now that we have the DB loaded, lets crate a slice
	var voterList []Vote
	var vtr Vote

	//Lets query redis for all of the items
	pattern := RedisKeyPrefix + "*"
	ks, _ := t.cacheClient.Keys(t.context, pattern).Result()
	for _, key := range ks {
		err := t.getVoteFromRedis(key, &vtr)
		if err != nil {
			return nil, err
		}
		voterList = append(voterList, vtr)
	}

	return voterList, nil
}
