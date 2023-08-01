package main

import (
	"voter-api/api"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
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
	flag.UintVar(&portFlag, "p", 1080, "Default Port")

	flag.Parse()
}


func main() {
	processCmdLineFlags()

	api := *api.NewVoterApi()

	r := gin.Default()
	r.GET("/voters", api.GetVotersApi)
	r.GET("/voters/:id", api.GetVoterApi)
	r.POST("/voters/:id", api.PostVoterApi)
	r.GET("/voters/:id/polls", api.GetVoterPollsApi)
	r.GET("/voters/:id/polls/:pollid", api.GetPollApi)
	r.POST("/voters/:id/polls/:pollid", api.PostPollApi)
	r.GET("/voters/health", api.GetHealthApi)

	serverPath := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	r.Run(serverPath)
}
