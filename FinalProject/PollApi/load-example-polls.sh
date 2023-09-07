
curl -d '{ "pollTitle": "Example poll 1", "pollQuestion": "Favorite food?", "pollOptions": ["Pizza", "Sandwich", "Bread"]}' -H "Content-Type: application/json" -X POST http://localhost:3080/poll
curl -d '{ "pollTitle": "Example poll 2", "pollQuestion": "Favorite TV Show?", "pollOptions": ["90210", "Ugly Betty", "Cops"]}' -H "Content-Type: application/json" -X POST http://localhost:3080/poll
curl -d '{ "pollTitle": "Example poll 3", "pollQuestion": "Favorite Movie?", "pollOptions": ["Barbie", "Oppenheimer", "Top Gun"]}' -H "Content-Type: application/json" -X POST http://localhost:3080/poll
curl -d '{ "pollTitle": "Example poll 4", "pollQuestion": "Favorite Place?", "pollOptions": ["NY", "LA", "PHL"]}' -H "Content-Type: application/json" -X POST http://localhost:3080/poll
