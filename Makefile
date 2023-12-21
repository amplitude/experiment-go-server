xpmt: clean
	go build -o xpmt cmd/xpmt/main.go

docker:
	docker build -t experiment . --progress plain
	docker run -it --rm --name experiment-run experiment

clean:
	rm -f xpmt

all: darwin linux

darwin: darwin-amd64 darwin-arm64

linux: linux-amd64 linux-arm64

darwin-amd64:
	go build -o cmd/xpmt/bin/darwin/amd64/xpmt cmd/xpmt/main.go

darwin-arm64:
	go build -o cmd/xpmt/bin/darwin/arm64/xpmt cmd/xpmt/main.go

linux-amd64:
	go build -o cmd/xpmt/bin/linux/amd64/xpmt cmd/xpmt/main.go

linux-arm64:
	go build -o cmd/xpmt/bin/linux/arm64/xpmt cmd/xpmt/main.go
