GO=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go
TAG=v0.1.0
BIN=critic
BIN_PATH=build/$(BIN)
IMAGE=eugenelukin/k8s-worker-critic

all: image
	docker push $(IMAGE):$(TAG)

build: clean
	$(GO) build -o $(BIN_PATH) .

image: build
	docker build -t $(IMAGE):$(TAG) .

clean:
	rm -f $(BIN_PATH)
