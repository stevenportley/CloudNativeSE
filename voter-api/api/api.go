package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"voter-api/voter"
)

type VoterApi struct {
	voterList voter.VoterList
}

func NewVoterApi() *VoterApi {
	return &VoterApi{
		voterList: voter.VoterList{
			Voters: make(map[uint]voter.Voter),
		},
	}
}

func (v *VoterApi) AddVoter(voterID uint, firstName, lastName string) {
	v.voterList.Voters[voterID] = *voter.NewVoter(voterID, firstName, lastName)
}

func (v *VoterApi) AddPoll(voterID, pollID uint) {
	voter := v.voterList.Voters[voterID]
	voter.AddPoll(pollID)
	v.voterList.Voters[voterID] = voter
}

func (v *VoterApi) GetVoter(voterID uint) voter.Voter {
	voter := v.voterList.Voters[voterID]
	return voter
}

func (v *VoterApi) GetVoterJson(voterID uint) string {
	voter := v.voterList.Voters[voterID]
	return voter.ToJson()
}

func (v *VoterApi) GetVoterList() voter.VoterList {
	return v.voterList
}

func (v *VoterApi) GetVoterListJson() string {
	b, _ := json.Marshal(v.voterList)
	return string(b)
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

	vtr := v.GetVoter(uint(id64))
	c.JSON(http.StatusOK, vtr)
}

func (v *VoterApi) PostVoterApi(c *gin.Context) {
	var voter voter.Voter

	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	v.AddVoter(voter.VoterID, voter.FirstName, voter.LastName)
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

	vtr := v.GetVoter(uint(id64))
	c.JSON(http.StatusOK, vtr.VoteHistory)
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

	vtr := v.GetVoter(uint(id64))
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

	vtr := v.GetVoter(uint(id64))

	poll := vtr.VoteHistory[pid]
	v.AddPoll(vtr.VoterID, poll.PollID)
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
