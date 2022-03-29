

xpmt: clean
	go build -o xpmt cmd/xpmt/main.go
	GOOS=darwin GOARCH=amd64 go build -o cmd/xpmt/bin/amd64/xpmt cmd/xpmt/main.go
	GOOS=darwin GOARCH=arm64 go build -o cmd/xpmt/bin/arm64/xpmt cmd/xpmt/main.go

clean:
	rm -f xpmt
