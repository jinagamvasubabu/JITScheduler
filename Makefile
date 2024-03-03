run:
	go run main.go

build-docker: build
	docker build . -t JITScheduler

run-docker: build-docker
	docker run -p 3000:3000 JITScheduler