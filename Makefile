genProto:
	cd server/internal/ports/grpcserver/
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	export PATH="$PATH:$(go env GOPATH)/bin"
	go generate ./proto
	cd ../../../../

BuildAndRun: genProto runServer
run: genProto runRedis runServer

runRedis:
	docker run --name redisSound -p 8088:6379 -d redis
	docker run --name redisUserInfo -p 8089:6379 -d redis

runServer:
	go run server/cmd/main.go

runTests:
	go test ./...

stopRedis:
	docker stop redisSound
	docker rm redisSound

	docker stop redisUserInfo
	docker rm redisUserInfo