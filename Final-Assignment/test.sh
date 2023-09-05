# Load DB with Polls, Voters, Votes

#create poll
echo "Sending create poll 1 request.\n"
curl -d '{"PollID": 1,  "PollTitle": "Favorite Color", "PollQuestion": "What is your favorite color?"}' -X POST "http://localhost:1082/polls/1"
echo
echo
echo "Sending create poll option 1 request.\n"
curl -d '{"PollOptionID": 1, "PollOptionText": "Blue"}' -X POST "http://localhost:1082/polls/1/polloption/1"  
echo
echo
echo "Sending create poll option 2 request.\n"
curl -d '{"PollOptionID": 2, "PollOptionText": "Brown"}' -X POST "http://localhost:1082/polls/1/polloption/2"  

#create voters
echo
echo
echo "Sending create voter 1 request.\n"
curl -d '{"VoterID": 1,"FirstName": "Michael","LastName": "Dratch"}' -X POST "http://localhost:1081/voters/1"
echo
echo
echo "Sending create voter 2 request.\n"
curl -d '{"VoterID": 2,"FirstName": "Bob","LastName": "Dylan"}' -X POST "http://localhost:1081/voters/2"
echo
echo
echo "Sending create voter 3 request.\n"
curl -d '{"VoterID": 3,"FirstName": "Rocky","LastName": "Balboa"}' -X POST "http://localhost:1081/voters/3"

#create votes
echo
echo
echo "Sending create vote 1 request.\n"
curl -d '{"VoteID": 1,"VoterID": 1,"PollID": 1,"PollOptionID": 1}' -X POST "http://localhost:1080/votes/1"
echo
echo
echo "Sending create vote 2 request.\n"
curl -d '{"VoteID": 2,"VoterID": 2,"PollID": 1,"PollOptionID": 2}' -X POST "http://localhost:1080/votes/2"
echo
echo
echo "Sending create vote 3 request.\n"
curl -d '{"VoteID": 3,"VoterID": 3,"PollID": 1,"PollOptionID": 1}' -X POST "http://localhost:1080/votes/3"

# Get requests with hypermedia

echo
echo
echo "Fetching Vote 1\n"
curl -X GET "http://localhost:1080/votes/1"
echo
echo
echo "Fetching Vote 2\n"
curl -X GET "http://localhost:1080/votes/2"
echo
echo
echo "Fetching Vote 3\n"
curl -X GET "http://localhost:1080/votes/3"

# Get requests with details provides through inter service commuication

echo
echo
echo "Fetching Vote 1 with details\n"
curl -X GET "http://localhost:1080/votes/1?detail=true"
echo
echo
echo "Fetching Vote 2 with details\n"
curl -X GET "http://localhost:1080/votes/2?detail=true"
echo
echo
echo "Fetching Vote 3 with details\n"
curl -X GET "http://localhost:1080/votes/3?detail=true"