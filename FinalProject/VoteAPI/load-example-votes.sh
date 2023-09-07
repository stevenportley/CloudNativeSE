
curl -d '{ "voterID": 1, "pollID": 1, "voteValue": 1}' -H "Content-Type: application/json" -X POST http://localhost:1080/vote
curl -d '{ "voterID": 1, "pollID": 2, "voteValue": 1}' -H "Content-Type: application/json" -X POST http://localhost:1080/vote
curl -d '{ "voterID": 2, "pollID": 1, "voteValue": 1}' -H "Content-Type: application/json" -X POST http://localhost:1080/vote
curl -d '{ "voterID": 2, "pollID": 2, "voteValue": 2}' -H "Content-Type: application/json" -X POST http://localhost:1080/vote
curl -d '{ "voterID": 3, "pollID": 1, "voteValue": 2}' -H "Content-Type: application/json" -X POST http://localhost:1080/vote
