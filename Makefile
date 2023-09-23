APP          := s3scanner
DOCKER_IMAGE := hothamandcheese/s3scanner
VERSION      := $(shell git describe --tags --abbrev=0)
COMMIT       := $(shell git rev-parse --short HEAD)
BUILD_DATE   := `date +%FT%T%z`

dev:
	docker compose -f .dev/docker-compose.yml --profile default up -d

dev-mitm:
	docker compose -f .dev/docker-compose.yml --profile dev-mitm up -d

docker-image:
	docker build -t $(DOCKER_IMAGE):$(VERSION) -f packaging/docker/Dockerfile .

lint:
	docker run -t --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.53.3 golangci-lint run -v 

test:
	go test ./...

test-coverage:
	TEST_DB=1 TEST_MQ=1 go test ./... -coverprofile cover.out
	go tool cover -html=cover.out

test-integration:
	TEST_DB=1 TEST_MQ=1 go test ./...

upgrade:
	go get -u ./...
