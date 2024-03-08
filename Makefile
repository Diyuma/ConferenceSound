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

runServer:
	go run server/cmd/main.go

runTests:
	go test ./...

stopRedis:
	docker stop redisSound
	docker rm redisSound

	docker stop redisUserInfo
	docker rm redisUserInfo


BuildAndPushDockerImage:
	docker build . --tag lehatr/conferencesoundserver
	docker push lehatr/conferencesoundserver

# ------- SERVER TARGETS -------
# don't forget to add commands to change ips there
uploadToServer:
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 rm -rf "~/conference/soundServer/"
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 rm -rf "~/conference/restServer/"

	rm -rf soundServer
	mkdir -p soundServer

	cp -r server soundServer/server
	cp -r restServer soundServer/restServer

	cp Makefile envoy-override.yaml soundServer

	scp -i ~/.ssh/yconference -r soundServer lehatr@178.154.202.56:~/conference

	scp -i ~/.ssh/yconference nginx.conf lehatr@178.154.202.56:~/docker-nginx/nginx.conf

	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker build "~/conference/soundServer/server" --tag lehatr/conferencesoundserver
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker build "~/conference/soundServer/restServer" --tag lehatr/conferencerestserver
	rm -rf soundServer

runOnServer:
	-ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker network rm -f conf_net
	-ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker network create --subnet=172.18.0.0/16 conf_net

	-ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker stop redisSound
	-ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker rm redisSound
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker run --name redisSound --network conf_net --ip 172.18.0.5 -d redis


	-ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker stop Envoy
	-ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker rm Envoy

	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker run --name Envoy --network conf_net --publish 8085:8085 -d \
				-v /home/lehatr/conference/soundServer/envoy-override.yaml:/envoy-override.yaml \
				envoyproxy/envoy-dev:c11574972860a40de36acf3ab8d930273f5ece65 \
				-c /envoy-override.yaml

	-ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker stop soundServer
	-ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker rm soundServer
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker run --name soundServer --network conf_net --ip 172.18.0.6 -d \
				lehatr/conferencesoundserver

	-ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker stop frontFileServer
	-ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker rm frontFileServer
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker run --name frontFileServer --publish 8086:8086 -d lehatr/conferencerestserver

	-ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker stop docker-nginx
	-ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker rm docker-nginx
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker run --name docker-nginx -p 443:443 --network conf_net -v /home/lehatr/conference/html:/usr/share/nginx/html \
				-v /home/lehatr/docker-nginx/nginx.conf:/etc/nginx/conf.d/default.conf nginx

connectToServer:
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56

getLogsFromServer:
	-rm -r logs
	-ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo rm -r "~/conference/logs"
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 mkdir -p "~/conference/logs"

	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker cp soundServer:/app/sound_loggs.log "~/conference/logs/sound_loggs.log"
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo docker cp soundServer:/app/repo_loggs.log "~/conference/logs/repo_loggs.log"
	
	scp -i ~/.ssh/yconference -r lehatr@178.154.202.56:~/conference/logs .
	ssh -i ~/.ssh/yconference lehatr@178.154.202.56 sudo rm -r "~/conference/logs"

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
	docker run --name soundServer --network conf_net --ip 172.18.0.6 lehatr/conferencesoundserver

stopServerServerSide:
	docker stop soundServer
	docker rm soundServer

stopRestServerServerSide:
	docker stop frontFileServer
	docker rm frontFileServer

# TODO rewrite without pwd=/home/lehatr/conference/soundServer
runEnvoyServerSide:
	docker run --network conf_net --publish 8085:8085 -v /home/lehatr/conference/soundServer/envoy-override.yaml:/envoy-override.yaml \
				envoyproxy/envoy-dev:c11574972860a40de36acf3ab8d930273f5ece65 \
				-c /envoy-override.yaml



### depracated go
# make
# docker
### depreacted envoy
### deprecated and configure nginx:)))

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