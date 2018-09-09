run:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o go-app -a .
	docker build -t go-app:1.0.0 .
	docker-compose up

test:
	go test -v .
