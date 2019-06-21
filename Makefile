# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=cmduptodate
PROJECT_ROOT=github.com/abigpotostew/$(BINARY_NAME)
BUILD_DIR=bin
BINARY_OUTPUT=$(BUILD_DIR)/$(BINARY_NAME)

all: clean build

build:
		$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v

clean:
		$(GOCLEAN)
		rm -rf $(BUILD_DIR)

run: build
		$(BINARY_OUTPUT) -c $(PROJECT_ROOT) -g $(BINARY_OUTPUT)