package api

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"vote-api/vote"
)

type VoteApi struct {
	voterDb vote.VoteDb
}

func NewVoteApi() (*VoteApi, error) {

	db, err := vote.NewDb()

	if err != nil {
		log.Println("Failed to initialize the Voter database, is redis running?")
		return nil, err
	}

	return &VoteApi{
		voterDb: *db,
	}, nil
}

func (v *VoteApi) AddVote(voteID uint, voterID uint, pollID uint, voteValue uint) {
	err := v.voterDb.AddVote(*vote.NewVote(voteID, voterID, pollID, voteValue))
	if err != nil {
		log.Println("Failed to add voter to the database: ", err)
	}
}

func (v *VoteApi) GetVote(voteID int) (vote.Vote, error) {
	vtr, err := v.voterDb.GetVote(int(voteID))

	if err != nil {
		log.Println("Failed to get a voter, does not exist in db")
		return vote.Vote{}, err
	}

	return vtr, nil
}

func (v *VoteApi) GetVoteList() []vote.Vote {
	vtrs, err := v.voterDb.GetAllVotes()
	if err != nil {
		log.Println("Failed to fetch the list of voters from the database!")
	}

	// If GetAllVotes() fails, return the empty list anyway
	return vtrs
}

func (v *VoteApi) GetVotersApi(c *gin.Context) {
	voters := v.GetVoteList()
	c.JSON(http.StatusOK, voters)
}

func (v *VoteApi) GetVoteApi(c *gin.Context) {
	id := c.Param("id")
	id64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		log.Println("Error converting id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	vtr, err := v.GetVote(int(id64))
	if err != nil {
		log.Println("Failed to fetch a voter from the DB!")
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, vtr)
	}
}

func (v *VoteApi) PostVoterApi(c *gin.Context) {
	var voter vote.Vote

	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id := c.Param("id")
	id64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		log.Println("Error converting id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	v.AddVote(uint(id64), voter.FirstName, voter.LastName)
	c.JSON(http.StatusOK, voter)
}

func (v *VoteApi) GetHealthApi(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":             "ok",
		"version":            "1.0.0",
		"uptime":             100,
		"users_processed":    1000,
		"errors_encountered": 10,
	})
}
