.PHONY: build

build:
	go build -o . ./...
build-docker:
	docker build -f ./docker/Dockerfile -t streamsculpt .
run-docker:
	docker run -it --rm --env-file=.env streamsculpt
format:
	gofumpt -l -w .
