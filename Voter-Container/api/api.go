package api

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"voter-api/voter"
)

type VoterApi struct {
	voterDb voter.VoterDb
}

func NewVoterApi() (*VoterApi, error) {

	db, err := voter.NewDb()

	if err != nil {
		log.Println("Failed to initialize the Voter database, is redis running?")
		return nil, err
	}

	return &VoterApi{
		voterDb: *db,
	}, nil
}

func (v *VoterApi) AddVoter(voterID uint, firstName, lastName string) {
	err := v.voterDb.AddVoter(*voter.NewVoter(voterID, firstName, lastName))
	if err != nil {
		log.Println("Failed to add voter to the database: ", err)
	}
}

func (v *VoterApi) AddPoll(voterID, pollID uint) error {
	voter, err := v.voterDb.GetVoter(int(voterID))
	if err != nil {
		log.Println("Failed to add a poll because the requested voter does not exist!")
		return err
	}

	voter.AddPoll(pollID)

	v.voterDb.UpdateVoter(voter)
	return nil
}

func (v *VoterApi) GetVoter(voterID int) (voter.Voter, error) {
	vtr, err := v.voterDb.GetVoter(int(voterID))

	if err != nil {
		log.Println("Failed to get a voter, does not exist in db")
		return voter.Voter{}, err
	}

	return vtr, nil
}

func (v *VoterApi) GetVoterList() []voter.Voter {
	vtrs, err := v.voterDb.GetAllVoters()
	if err != nil {
		log.Println("Failed to fetch the list of voters from the database!")
	}

	// If GetAllVoters() fails, return the empty list anyway
	return vtrs
}

func (v *VoterApi) GetVotersApi(c *gin.Context) {
	voters := v.GetVoterList()
	c.JSON(http.StatusOK, voters)
}

func (v *VoterApi) GetVoterApi(c *gin.Context) {
	id := c.Param("id")
	id64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		log.Println("Error converting id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	vtr, err := v.GetVoter(int(id64))
	if err != nil {
		log.Println("Failed to fetch a voter from the DB!")
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, vtr)
	}
}

func (v *VoterApi) PostVoterApi(c *gin.Context) {
	var voter voter.Voter

	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	log.Println("Adding voter...")
	log.Println("ID: ", voter.VoterID)
	log.Println("FirstName ", voter.FirstName)
	log.Println("LastName ", voter.LastName)

	id := c.Param("id")
	id64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		log.Println("Error converting id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	v.AddVoter(uint(id64), voter.FirstName, voter.LastName)
	c.JSON(http.StatusOK, voter)
}

func (v *VoterApi) GetVoterPollsApi(c *gin.Context) {
	id := c.Param("id")
	id64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		log.Println("Error converting id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	vtr, err := v.GetVoter(int(id64))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, vtr.VoteHistory)
	}
}

func (v *VoterApi) GetPollApi(c *gin.Context) {
	id := c.Param("id")
	id64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		log.Println("Error converting id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	pollid := c.Param("pollid")
	pid, err := strconv.ParseUint(pollid, 10, 32)
	if err != nil {
		log.Println("Error converting id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	vtr, err := v.GetVoter(int(id64))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	poll := vtr.VoteHistory[pid]
	c.JSON(http.StatusOK, poll)
}

func (v *VoterApi) PostPollApi(c *gin.Context) {
	id := c.Param("id")
	id64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		log.Println("Error converting id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	pollid := c.Param("pollid")
	pid, err := strconv.ParseUint(pollid, 10, 32)
	if err != nil {
		log.Println("Error converting id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	vtr, err := v.GetVoter(int(id64))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	err = v.AddPoll(vtr.VoterID, uint(pid))
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, pid)
}

func (v *VoterApi) GetHealthApi(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":             "ok",
		"version":            "1.0.0",
		"uptime":             100,
		"users_processed":    1000,
		"errors_encountered": 10,
	})
}
