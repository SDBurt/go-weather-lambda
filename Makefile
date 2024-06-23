.PHONY: build clean

build:
	GOOS=linux GOARCH=arm64 go build -o bootstrap main.go
	zip lambda-handler.zip bootstrap

clean:
	rm -f bootstrap lambda-handler.zip