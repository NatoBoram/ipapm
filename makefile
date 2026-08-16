build:
	CGO_ENABLED=0 go build -trimpath

clean:
	rm -rf ipapm ipapm.exe root vendor

format:
	go tool gofumpt -extra -w .

lint:
	go tool golangci-lint run

prod: clean build
	GO_ENV=production ./ipapm

run:
	GO_ENV=development go run -trimpath .

livez:
	GO_ENV=development go run -trimpath ./cmd/livez

readyz:
	GO_ENV=development go run -trimpath ./cmd/readyz

test:
	GO_ENV=test go test -count=1 ./...

compose-build:
	docker compose build --pull

compose-up:
	docker compose up --detach --force-recreate --remove-orphans

compose-down:
	docker compose down --remove-orphans

docker-build:
	docker buildx build -t natoboram/ipapm .

docker-export:
	docker buildx build -o type=local,dest=./root -t natoboram/ipapm .
