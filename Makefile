.ONESHELL:
.SHELLFLAGS += -e

genProto:
	cd server/internal/ports/grpcserver/
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	export PATH="$PATH:$(go env GOPATH)/bin"
	go generate ./proto
	cd ../../../../

BuildAndRun: genProto runServer
run: genProto runRedis runServer

runEnvoy:
	envoy -c envoy-override.yaml

runRedis:
	docker run --name redisSound -p 6379:6379 -d redis
# docker run --name redisUserInfo -p 8089:6379 -d redis

runServer:
	go run server/cmd/main.go

runTests:
	go test ./...

stopRedis:
	docker stop redisSound
	docker rm redisSound

	docker stop redisUserInfo
	docker rm redisUserInfo

# docker login -u lehatrutenb@gmail.com
# --platform="linux/amd64"
BuildAndPushDockerImage:
	docker build . --tag lehatr/conferencesoundserver
	docker push lehatr/conferencesoundserver

# ------- SERVER TARGETS -------
# don't forget to add commands to change ips there
# don't forget to add sudos to make with docker ? or just to run make with sudo???
uploadToServer:
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 rm -rf "~/conference/soundServer/"

	rm -rf soundServer
	mkdir -p soundServer

	cp -r server soundServer/server
	cp main.go go.mod go.sum soundServer
	cp Makefile envoy-override.yaml Dockerfile soundServer

	scp -i ~/.ssh/yconference -r soundServer lehatr@178.154.202.56:~/conference/
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker build "~/conference/soundServer/" --tag lehatr/conferencesoundserver
	rm -rf soundServer


runOnServer:
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker network rm conf_net & sudo docker network create --subnet=172.18.0.0/16 conf_net

	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker stop redisSound & sudo docker rm redisSound
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker run --name redisSound --network conf_net --ip 172.18.0.5 -d redis


	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 docker run --network conf_net --publish 8085:8085 \
				-v /home/lehatr/conference/soundServer/envoy-override.yaml:/envoy-override.yaml \
				envoyproxy/envoy-dev:c11574972860a40de36acf3ab8d930273f5ece65 \
				-c /envoy-override.yaml


	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker stop lehatr/conferencesoundserver
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker run --network conf_net --ip 172.18.0.6 lehatr/conferencesoundserver


connectToServer:
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56

ctDockerNetwork:
	docker network create --subnet=172.18.0.0/16 conf_net

rmDockerNetwork:
	docker network rm conf_net

runRedisServerSide:
	docker run --name redisSound --network conf_net --ip 172.18.0.5 -d redis

stopRedisServerSide:
	docker stop redisSound
	docker rm redisSound

runServerServerSide:
	docker run --network conf_net --ip 172.18.0.6 lehatr/conferencesoundserver

# TODO rewrite without pwd=/home/lehatr/conference/soundServer
runEnvoyServerSide:
	docker run --network conf_net --publish 8085:8085 -v /home/lehatr/conference/soundServer/envoy-override.yaml:/envoy-override.yaml \
				envoyproxy/envoy-dev:c11574972860a40de36acf3ab8d930273f5ece65 \
				-c /envoy-override.yaml




# go // because now server not in docker
# make
# docker
# envoy

serverSetup:
	sudo apt-get update

	sudo apt search golang-go
	sudo apt search gccgo-go

	sudo apt install make

	sudo apt-get update
	sudo apt-get install ca-certificates curl
	sudo install -m 0755 -d /etc/apt/keyrings
	sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
	sudo chmod a+r /etc/apt/keyrings/docker.asc

	# Add the repository to Apt sources:
	echo \
	"deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
	$(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
	sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

	sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

	sudo docker pull envoyproxy/envoy-dev:c11574972860a40de36acf3ab8d930273f5ece65