build:
	CGO_ENABLED=0 go build -trimpath

clean:
	rm -rf ipapm ipapm.exe vendor

format:
	go tool gofumpt -extra -w .

lint:
	go tool golangci-lint run

prod: clean build
	GO_ENV=production ./ipapm

run:
	GO_ENV=development go run -trimpath ./...

test:
	GO_ENV=test go test -count=1 ./...

docker-build:
	docker buildx build -t ipapm .

docker-run:
	docker run ipapm

docker-kill:
	docker ps --format '{{.Image}} {{.ID}}' | grep ipapm | awk '{print $2}' | xargs -r docker kill
