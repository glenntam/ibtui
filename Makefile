.PHONY: all update build run clean

APP := ibtui

all: run clean

update:
	go get -u ./...
	go mod tidy

build:
	go build -o $(APP) ./cmd/tui/

run: build
	./$(APP)

clean:
	rm -f ./$(APP)
