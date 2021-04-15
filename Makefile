bootstrap: kloud
	go build -o bootstrapper -ldflags="-s -w" bootstrap/bootstrap.go

kloud:;
	GOOS=linux GOARCH=arm go build -o bootstrap/kloud -ldflags="-s -w" cmd/main.go

.PHONY: kloud bootstrap
