SHELL := /bin/bash

.PHONY: help
help:
	@echo "Usage make <TARGET>"
	@echo ""
	@echo "  Targets:"
	@echo "	   build				Build the todo executable"
	@echo "	   run					Run the todo program from code"
	@echo "	   load-db				Add sample data via curl"
	@echo "	   get-by-id			Get a todo by id pass id=<id> on command line"
	@echo "	   get-all				Get all todos"
	@echo "	   update-2				Update record 2, pass a new title in using title=<title> on command line"
	@echo "	   delete-all			Delete all todos"
	@echo "	   delete-by-id			Delete a todo by id pass id=<id> on command line"



.PHONY: run
run:
	go run main.go


.PHONY: load-db
load-db:
	curl -d '{ "id": 1, "title": "Learn Go / GoLang", "done": false }' -H "Content-Type: application/json" -X POST http://localhost:1080/todo 
	curl -d '{ "id": 2, "title": "Learn Kubernetes", "done": true}' -H "Content-Type: application/json" -X POST http://localhost:1080/todo 
	curl -d '{ "id": 3, "title": "Learn Cloud Native Architecture","done": false}' -H "Content-Type: application/json" -X POST http://localhost:1080/todo 
	curl -d '{"id": 4,"title": "Learn Why Professor Mitchell is the BEST! :-)","done": true}' -H "Content-Type: application/json" -X POST http://localhost:1080/todo

.PHONY: update-2
update-2:
	curl -d '{ "id": 2, "title": "$(title)", "done": false }' -H "Content-Type: application/json" -X PUT http://localhost:1080/todo 

.PHONY: get-by-id
get-by-id:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/todo/$(id) 

.PHONY: get-all
get-all:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/todo 

.PHONY: delete-all
delete-all:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X DELETE http://localhost:1080/todo 

.PHONY: delete-by-id
delete-by-id:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X DELETE http://localhost:1080/todo/$(id) 

