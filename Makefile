.PHONY: docker
docker:
	rm byline || true
	GOOS=linux GOARCH=amd64 go build -o byline .
	docker rmi -f jaylancharles/byline:v0.0.4 || true
	docker build -t jaylancharles/byline:v0.0.4 .