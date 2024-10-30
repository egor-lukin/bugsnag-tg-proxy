SHA := $(shell git rev-parse HEAD)

build:
	docker build -t bugsnag-tg-proxy:${SHA} .

test:
	go test
