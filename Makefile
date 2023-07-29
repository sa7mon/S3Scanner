APP          := s3scanner
DOCKER_IMAGE := hothamandcheese/s3scanner
VERSION      := $(shell git describe --tags --abbrev=0)
COMMIT       := $(shell git rev-parse --short HEAD)
BUILD_DATE   := `date +%FT%T%z`

dev:
	docker compose -f .dev/docker-compose.yml up -d

docker-image:
	docker build -t $(DOCKER_IMAGE):$(VERSION) -f packaging/docker/Dockerfile .

test:
	go test ./...

test-integration:
	TEST_DB=1 TEST_MQ=1 go test ./...
