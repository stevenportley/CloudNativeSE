package voter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
	"log"
	"os"
	"time"
)

type voterPoll struct {
	PollID   uint      `json:"pollid"`
	VoteDate time.Time `json:"VoteDate"`
}

type Voter struct {
	VoterID     uint   `json:"id"`
	FirstName   string `json:"FirstName"`
	LastName    string `json:"LastName"`
	VoteHistory []voterPoll
}
type VoterDb struct {
	cacheClient *redis.Client
	jsonHelper  *rejson.Handler
	context     context.Context
}

const (
	RedisNilError        = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisKeyPrefix       = "voter:"
)

// constructor for VoterList struct
func NewVoter(id uint, fn, ln string) *Voter {
	return &Voter{
		VoterID:     id,
		FirstName:   fn,
		LastName:    ln,
		VoteHistory: []voterPoll{},
	}
}

func (v *Voter) AddPoll(pollID uint) {
	v.VoteHistory = append(v.VoteHistory, voterPoll{PollID: pollID, VoteDate: time.Now()})
}

func NewDb() (*VoterDb, error) {
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}

	return NewDbWithInstance(redisUrl)
}

func NewDbWithInstance(location string) (*VoterDb, error) {

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
	return &VoterDb{
			cacheClient: client,
			jsonHelper:  jsonHelper,
			context:     ctx,
		},
		nil
}

//------------------------------------------------------------
// REDIS HELPERS
//------------------------------------------------------------

// We will use this later, you can ignore for now
func isRedisNilError(err error) bool {
	return errors.Is(err, redis.Nil) || err.Error() == RedisNilError
}

// In redis, our keys will be strings, they will look like
// todo:<number>.  This function will take an integer and
// return a string that can be used as a key in redis
func redisKeyFromId(id int) string {
	return fmt.Sprintf("%s%d", RedisKeyPrefix, id)
}

func (t *VoterDb) getItemFromRedis(key string, item *Voter) error {

	//Lets query redis for the item, note we can return parts of the
	//json structure, the second parameter "." means return the entire
	//json structure
	itemObject, err := t.jsonHelper.JSONGet(key, ".")
	if err != nil {
		log.Println("Failed to get JSON key object...")
		return err
	}

	//JSONGet returns an "any" object, or empty interface,
	//we need to convert it to a byte array, which is the
	//underlying type of the object, then we can unmarshal
	//it into our ToDoItem struct
	err = json.Unmarshal(itemObject.([]byte), item)
	if err != nil {
		log.Println("Failed to un-marshal voter object...")
		return err
	}

	return nil
}

func (t *VoterDb) AddVoter(voter Voter) error {

	log.Println("Adding voter...")

	//Before we add an item to the DB, lets make sure
	//it does not exist, if it does, return an error
	redisKey := redisKeyFromId(int(voter.VoterID))
	var existingItem Voter
	log.Println("Redis key...", redisKey)
	if err := t.getItemFromRedis(redisKey, &existingItem); err == nil {
		return errors.New("voter already exists")
	}

	//Add item to database with JSON Set
	if _, err := t.jsonHelper.JSONSet(redisKey, ".", voter); err != nil {
		return err
	}

	//If everything is ok, return nil for the error
	return nil
}

func (t *VoterDb) GetVoter(id int) (Voter, error) {

	// Check if item exists before trying to get it
	// this is a good practice, return an error if the
	// item does not exist
	var voter Voter
	pattern := redisKeyFromId(id)
	err := t.getItemFromRedis(pattern, &voter)
	if err != nil {
		return Voter{}, err
	}

	return voter, nil
}

func (t *VoterDb) UpdateVoter(voter Voter) error {

	//Before we add an item to the DB, lets make sure
	//it does not exist, if it does, return an error
	redisKey := redisKeyFromId(int(voter.VoterID))
	var existingVoter Voter
	if err := t.getItemFromRedis(redisKey, &existingVoter); err != nil {
		return errors.New("voter does not exist")
	}

	//Add item to database with JSON Set.  Note there is no update
	//functionality, so we just overwrite the existing item
	if _, err := t.jsonHelper.JSONSet(redisKey, ".", voter); err != nil {
		return err
	}

	//If everything is ok, return nil for the error
	return nil
}

func (t *VoterDb) GetAllVoters() ([]Voter, error) {

	//Now that we have the DB loaded, lets crate a slice
	var voterList []Voter
	var vtr Voter

	//Lets query redis for all of the items
	pattern := RedisKeyPrefix + "*"
	ks, _ := t.cacheClient.Keys(t.context, pattern).Result()
	for _, key := range ks {
		err := t.getItemFromRedis(key, &vtr)
		if err != nil {
			return nil, err
		}
		voterList = append(voterList, vtr)
	}

	return voterList, nil
}
