

xpmt: clean
	go build -o xpmt cmd/xpmt/main.go

clean:
	rm xpmt