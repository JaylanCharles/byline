.PHONY: docker
docker:
	rm byline || true
	GOOS=linux GOARCH=amd64 go build -tags=k8s -o byline .
	docker rmi -f jaylancharles/byline:v0.0.4 || true
	docker build -t jaylancharles/byline:v0.0.5 .

.PHONY: mock
mock:
	@mockgen -source=internal/service/user.go -destination=internal/service/mocks/user.mock.go -package=svcmocks
	@mockgen -source=internal/service/code.go -destination=internal/service/mocks/code.mock.go -package=svcmocks
	@mockgen -source=internal/repository/user.go -destination=internal/repository/mocks/user.mock.go -package=repomocks
	@mockgen -source=internal/repository/code.go -destination=internal/repository/mocks/code.mock.go -package=repomocks
	@mockgen -source=internal/repository/dao/user.go -destination=internal/repository/dao/mocks/user.mock.go -package=daomocks
	@mockgen -source=internal/repository/cache/user.go -destination=internal/repository/cache/mocks/user.mock.go -package=cachemocks
	@go mod tidy