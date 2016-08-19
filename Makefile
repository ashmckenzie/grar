SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY=grar

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES) darwin

all: $(SOURCES) darwin

linux: $(SOURCES)
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -o bin/${BINARY}_linux_x86_64 .

darwin: $(SOURCES)
	mkdir -p bin
	CGO_ENABLED=0 GOOS=darwin go build -a -o bin/${BINARY}_darwin_x86_64 .

.PHONY: run
run:
	go run main.go $(filter-out $@, $(MAKECMDGOALS))

.PHONY: clean
clean:
	rm -f bin/*
