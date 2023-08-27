
curl -d '{ "FirstName": "Steven", "LastName": "Portley", "VoteHistory" : [] }' -H "Content-Type: application/json" -X POST http://localhost:1080/voters/6
curl -d '{ "FirstName": "ABCD", "LastName": "Portley", "VoteHistory" : [] }' -H "Content-Type: application/json" -X POST http://localhost:1080/voters/7
curl -d '{ "FirstName": "EFGH", "LastName": "Portley", "VoteHistory" : [] }' -H "Content-Type: application/json" -X POST http://localhost:1080/voters/8
curl -d '{ "FirstName": "IJKL", "LastName": "Portley", "VoteHistory" : [] }' -H "Content-Type: application/json" -X POST http://localhost:1080/voters/9
curl -d '{ "FirstName": "MNOP", "LastName": "Portley", "VoteHistory" : [] }' -H "Content-Type: application/json" -X POST http://localhost:1080/voters/10
curl -d '{ "FirstName": "QRST", "LastName": "Portley", "VoteHistory" : [] }' -H "Content-Type: application/json" -X POST http://localhost:1080/voters/11



curl -d '{}' -H "Content-Type: application/json" -X POST http://localhost:1080/voters/11/polls/1
