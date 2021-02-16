REPO=marcoamador
NAME=pistache
VERSION=1.1.4
INIT_VERSION=1.1.4

all: docker
clean: docker-clean

run:
	go run github.com/mfamador/pistache/cmd/pistache

build:
	go build github.com/mfamador/pistache/cmd/pistache

deps:
	go mod download

swag:
	swag init -g ./cmd/pistache/main.go -o ./doc/swagger

docker:
	docker build -f build/Dockerfile -t $(REPO)/$(NAME):$(VERSION) .
	docker build -f build/init/Dockerfile -t $(REPO)/$(NAME)-init:$(INIT_VERSION) .

docker-push:
	docker push $(REPO)/$(NAME):$(VERSION)
	docker push $(REPO)/$(NAME)-init:$(INIT_VERSION)

docker-clean:
	docker rmi $(REPO)/$(NAME):$(VERSION)

docker-run:
	docker run -f build/Dockerfile --rm --name $(NAME) $(REPO)/$(NAME):$(VERSION)

docker-test:
	docker build -f build/Dockerfile --target tester .

lint:
	docker run --rm \
		-w /app \
		-v $(shell pwd):/app:ro \
		-v $(shell pwd)/.cache/go-build:/root/.cache/go-build:rw \
		-v $(shell pwd)/.cache/golangci-lint:/root/.cache/golangci-lint:rw \
		-v $(shell pwd)/.cache/go/pkg/mod:/root/go/pkg/mod:rw \
		golangci-lint

.PHONY: build run lint
