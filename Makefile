bootstrap: kloud
	GOOS=windows GOARCH=amd64 go build -o bootstrapper_win_amd64 -ldflags="-s -w" bootstrap/bootstrap.go
	GOOS=linux GOARCH=amd64 go build -o bootstrapper_lin_amd64 -ldflags="-s -w" bootstrap/bootstrap.go

kloud:;
	GOOS=linux GOARCH=arm go build -o bootstrap/kloud -ldflags="-s -w" cmd/main.go

.PHONY: kloud bootstrap
