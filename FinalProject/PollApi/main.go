package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

var (
	hostFlag string
	portFlag uint
)

func processCmdLineFlags() {

	//Note some networking lingo, some frameworks start the server on localhost
	//this is a local-only interface and is fine for testing but its not accessible
	//from other machines.  To make the server accessible from other machines, we
	//need to listen on an interface, that could be an IP address, but modern
	//cloud servers may have multiple network interfaces for scale.  With TCP/IP
	//the address 0.0.0.0 instructs the network stack to listen on all interfaces
	//We set this up as a flag so that we can overwrite it on the command line if
	//needed
	flag.StringVar(&hostFlag, "h", "0.0.0.0", "Listen on all interfaces")
	flag.UintVar(&portFlag, "p", 3080, "Default Port")

	flag.Parse()
}

func main() {
	processCmdLineFlags()

	api, err := NewPollApi()

	if err != nil {
		log.Println("Failed to initialize the API... ", err)
		return
	}

	r := gin.Default()

	r.GET("/poll", func(c *gin.Context) {
		votes, err := api.GetAllPolls()
		if err != nil {
			log.Println("Failed to get votes from redis...", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, votes)
	})

	r.POST("/poll", func(c *gin.Context) {

		type Poll struct {
			PollTitle    string   `json:"pollTitle"`
			PollQuestion string   `json:"pollQuestion"`
			PollOptions  []string `json:"pollOptions"`
		}
		var poll Poll

		err := c.ShouldBindJSON(&poll)
		if err != nil {
			log.Println("Cannot fetch JSON body from vote POST", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		newPoll, err := api.AddPoll(poll.PollTitle, poll.PollQuestion, poll.PollOptions)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, newPoll)
	})

	r.GET("/poll/:id", func(c *gin.Context) {
		id := c.Param("id")
		id64, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			log.Println("Error converting id to int64: ", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		poll, err := api.GetPoll(int(id64))
		if err != nil {
			log.Println("Failed to get poll from redis...", err)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.JSON(http.StatusOK, poll)
	})

	// Hardcoded health status
	r.GET("/poll/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":             "ok",
			"version":            "1.0.0",
			"uptime":             100,
			"users_processed":    1000,
			"errors_encountered": 10,
		})
	})

	serverPath := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	r.Run(serverPath)
}
