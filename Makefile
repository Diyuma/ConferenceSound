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
	ssh -i &&KEY_PATH $$VM_USER@&&HOST rm -rf "~/conference/soundServer/"
	ssh -i &&KEY_PATH $$VM_USER@&&HOST rm -rf "~/conference/restServer/"

	rm -rf soundServer
	mkdir -p soundServer

	cp -r server soundServer/server
	cp -r restServer soundServer/restServer

	cp Makefile envoy-override.yaml soundServer

	scp -i &&KEY_PATH -r soundServer $$VM_USER@&&HOST:~/conference

	scp -i &&KEY_PATH nginx.conf $$VM_USER@&&HOST:~/docker-nginx/nginx.conf

	ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker build "~/conference/soundServer/server" --tag lehatr/conferencesoundserver
	ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker build "~/conference/soundServer/restServer" --tag lehatr/conferencerestserver
	rm -rf soundServer

runOnServer:
	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker network rm -f conf_net
	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker network create --subnet=172.18.0.0/16 conf_net

	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker stop redisSound
	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker rm redisSound
	ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker run --name redisSound --network conf_net --ip 172.18.0.5 -d redis

	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker stop redisInfo
	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker rm redisInfo
	ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker run --name redisInfo --network conf_net --ip 172.18.0.8 -d redis

	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker stop Envoy
	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker rm Envoy

	ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker run --name Envoy --network conf_net --publish 8085:8085 -d \
				-v /home/$$VM_USER/conference/soundServer/envoy-override.yaml:/envoy-override.yaml \
				envoyproxy/envoy-dev:c11574972860a40de36acf3ab8d930273f5ece65 \
				-c /envoy-override.yaml

	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker stop soundServer
	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker rm soundServer
	ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker run --name soundServer --network conf_net --ip 172.18.0.6 -d \
				lehatr/conferencesoundserver

	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker stop frontFileServer
	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker rm frontFileServer
	ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker run --name frontFileServer --publish 8086:8086 -d lehatr/conferencerestserver

	-ssh -i $$KEY_PATH $$VM_USER@$$HOST sudo docker stop videoServer
	-ssh -i $$KEY_PATH $$VM_USER@$$HOST sudo docker rm videoServer
	ssh -i $$KEY_PATH $$VM_USER@$$HOST sudo docker run --name videoServer --network conf_net --ip 172.18.0.7 -d \
				dikiray/conferencevideoserver

	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker stop docker-nginx
	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker rm docker-nginx
	ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker run --name docker-nginx -p 443:443 --network conf_net -d \
				-v /home/$$VM_USER/conference/html:/usr/share/nginx/html \
				-v /home/$$VM_USER/docker-nginx/nginx.conf:/etc/nginx/conf.d/default.conf nginx

connectToServer:
	ssh -i &&KEY_PATH $$VM_USER@&&HOST

getLogsFromServer:
	-rm -r logs
	-ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo rm -r "~/conference/logs"
	ssh -i &&KEY_PATH $$VM_USER@&&HOST mkdir -p "~/conference/logs"

	ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker cp soundServer:/app/sound_loggs.log "~/conference/logs/sound_loggs.log"
	ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo docker cp soundServer:/app/repo_loggs.log "~/conference/logs/repo_loggs.log"
	
	scp -i &&KEY_PATH -r $$VM_USER@&&HOST:~/conference/logs .
	ssh -i &&KEY_PATH $$VM_USER@&&HOST sudo rm -r "~/conference/logs"

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



# make
# docker
# docker nginx
# docker envoy
# docker redis
# need to run on server!

serverSetup:
	sudo apt-get update

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
	sudo docker pull nginx
	sudo docker pull redis

	sudo mkdir -p "~/conference/html/conference/dist/"
