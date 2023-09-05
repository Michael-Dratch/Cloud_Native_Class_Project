Michael Dratch \
9/4/23 \
CS 680

# Cloud Native Software Engineering Final Project

## Summary:

An Application for tracking polling data built out of three seperate containerized services:

- A Polls API that handles polls records
- A Voter API that handles Voter records
- And a Votes API that contains vote records and communicates with
  the other services to create integrated reports

## Instructions:

This application uses docker-compose to build and run all of the required docker containers.

- Run "docker-compose up" (must be in the Final-Assignment directory)
- Run the test.sh script which contains API calls to demonstrate creating polls, voters, and vote records and querying this data with various levels of detail

## Design Notes

My main use of hypermedia was in the vote api. A standard GET request for a vote record returns a JSON object where the fields are URL links which can be used to get further details. GETing a vote with the "detail" parameter set to true returns a larger JSON response where all of the details of the voter and poll related to the vote are provided. The vote API uses the hypermedia links in the Vote object to communicate with the other services and collect this information for the response. There is little hypermedia involved in creating records because It seems to me that these requests would likely come from voting applications where the interactions with the APIs are hardcoded and less flexible.
