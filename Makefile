GOBUILD=go build .
DOCKER_USER=stevelacy
DOCKER_TAG=latest
NAME=kubecat

all: docker

build:
	$(GOBUILD)

build_linux:
	GOOS=linux GOARCH=386 CGO_ENABLED=0 $(GOBUILD)

docker:
	docker build -t $(DOCKER_USER)/$(NAME):$(DOCKER_TAG) .

clean:
	rm -f $(NAME)