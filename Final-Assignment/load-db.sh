curl -d '{"PollID": 1,  "PollTitle": "Favorite Color", "PollQuestion": "What is your favorite color?"}' -X POST "http://localhost:1082/polls/1"
curl -d '{"PollID": 2,  "PollTitle": "Favorite Food", "PollQuestion": "What is your favorite Food?"}' -X POST "http://localhost:1082/polls/2" 

curl -d '{"PollOptionID": 1, "PollOptionText": "Blue"}' -X POST "http://localhost:1082/polls/1/polloption/1"  
curl -d '{"PollOptionID": 2, "PollOptionText": "Brown"}' -X POST "http://localhost:1082/polls/1/polloption/2"  

curl -d '{"VoterID": 1,"FirstName": "Michael","LastName": "Dratch"}' -X POST "http://localhost:1081/voters/1"
curl -d '{"VoterID": 2,"FirstName": "Bob","LastName": "Dylan"}' -X POST "http://localhost:1081/voters/2"

curl -d '{"VoteID": 1,"VoterID": 1,"PollID": 1,"PollOptionID": 1}' -X POST "http://localhost:1080/votes/1"
curl -d '{"VoteID": 2,"VoterID": 2,"PollID": 1,"PollOptionID": 2}' -X POST "http://localhost:1080/votes/2"