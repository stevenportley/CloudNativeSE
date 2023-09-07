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
	flag.UintVar(&portFlag, "p", 2080, "Default Port")

	flag.Parse()
}

func main() {
	processCmdLineFlags()

	api, err := NewVoterApi()

	if err != nil {
		log.Println("Failed to initialize the API... ", err)
		return
	}

	r := gin.Default()

	r.GET("/voter", func(c *gin.Context) {
		voters, err := api.GetAllVoters()
		if err != nil {
			log.Println("Failed to fetch all of the voters!", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, voters)
		return
	})

	r.POST("/voter", func(c *gin.Context) {

		type Voter struct {
			FirstName string `json:"FirstName"`
			LastName  string `json:"LastName"`
		}

		var voter Voter

		err := c.ShouldBindJSON(&voter)
		if err != nil {
			log.Println("Cannot fetch JSON body from voter POST", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		newVoter, err := api.AddVoter(voter.FirstName, voter.LastName)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		} else {
			c.JSON(http.StatusOK, newVoter)
		}
	})

	r.GET("/voter/:id", func(c *gin.Context) {
		id := c.Param("id")
		id64, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			log.Println("Error converting id to int64: ", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		vtr, err := api.GetVoter(int(id64))
		if err != nil {
			log.Println("Failed to fetch a voter from the DB!")
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			c.JSON(http.StatusOK, *vtr)
		}
	})

	r.POST("/voter/:id/:pollid", func(c *gin.Context) {
		id := c.Param("id")
		id64, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			log.Println("Error converting id to int64: ", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		pid := c.Param("pollid")
		pid64, err := strconv.ParseUint(pid, 10, 32)
		if err != nil {
			log.Println("Error converting id to int64: ", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		err = api.Vote(int(id64), uint(pid64))
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
		}

		c.JSON(http.StatusOK, gin.H{})
	})

	// Hardcoded health status
	r.GET("/voter/health", func(c *gin.Context) {
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
