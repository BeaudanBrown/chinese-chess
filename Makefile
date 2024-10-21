BINARY_NAME=main
DIST_FOLDER=./dist

.PHONY: build

run:
	go run cmd/main.go

cleandb:
	rm -rf .devenv/state/mongodb
	mv data/state* data/state.json

build:
	@echo "Building static binary..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(DIST_FOLDER)/$(BINARY_NAME) ./cmd/main.go

docker-build:
	@echo "Building dockerfile..."
	docker-compose build

docker-run: build
	@echo "Building dockerfile..."
	docker-compose build
	docker-compose up

docker-push: build
	@echo "Pushing docker image..."
	docker build -t mariahs-memories .
	docker tag mariahs-memories docker.beaudan.me/mariahs-memories
	docker push docker.beaudan.me/mariahs-memories:latest

.PHONY: clean

clean:
	@echo "Cleaning up..."
	rm -rf $(DIST_FOLDER)/$(BINARY_NAME)
