.ONESHELL:
.SHELLFLAGS += -e
include .bashrc

genProto:
	cd server/internal/ports/grpcserver/
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	export PATH="${PATH}:$(go env GOPATH)\bin"
	go generate ./proto
	cd ../../../../

buildAndRun:
	docker compose -p conference down
	docker compose build
	docker compose -p conference up

run:
	docker compose -p conference down
	docker compose -p conference up

stop:
	docker compose -p conference down

getLogs:
	mkdir -p logs
	docker cp conference-sound_server-1:/app/sound_loggs.log "./logs/sound_loggs.log"
	docker cp conference-sound_server-1:/app/repo_loggs.log "./logs/repo_loggs.log"

# ------- SERVER TARGETS -------
# it's not expected that they will work both on local machine and remote ubuntu in same way - so care!
uploadToServerPrepareDir:
	ssh -i ${KEY_PATH} ${VM_USER}@${HOST} sudo rm -rf "~/conferencev2/{soundServer/,restServer/,nginx.conf,envoy-override.yaml,.bashrc,go.work,go.work.sum,Makefile,ssl.conf,compose.yaml}"

uploadToServerNginxAndSslConf:
	cp nginx.conf nginx_server_generated.conf
	sed -i '' 's/listen 443;/listen 443 ssl;/g' nginx_server_generated.conf
	scp -i ${SSH_KEY_PATH} nginx_server_generated.conf ${VM_USER}@${HOST}:~/conferencev2/nginx.conf

uploadAndRunServer: uploadToServerPrepareDir uploadToServerNginxAndSslConf
	scp -i ${SSH_KEY_PATH} -r restServer server ${VM_USER}@${HOST}:~/conferencev2

	scp -i ${SSH_KEY_PATH} envoy-override.yaml go.work go.work.sum Makefile compose.yaml  ${VM_USER}@${HOST}:~/conferencev2
	scp -i ${SSH_KEY_PATH} ssl_server.conf ${VM_USER}@${HOST}:~/conferencev2/ssl.conf
	scp -i ${SSH_KEY_PATH} .bashrc_server ${VM_USER}@${HOST}:~/conferencev2/.bashrc

	ssh -i ${KEY_PATH} ${VM_USER}@${HOST} "cd conferencev2 && sudo make buildAndRun"

runServer:
	ssh -i ${KEY_PATH} ${VM_USER}@${HOST} "cd conferencev2 && sudo make run"

stopServer:
	ssh -i ${KEY_PATH} ${VM_USER}@${HOST} "cd conferencev2 && sudo make stop"

connectToServer:
	ssh -i ${SSH_KEY_PATH} ${VM_USER}@${HOST}

getLogsServer:
	-rm -r logs
	-ssh -i ${KEY_PATH} ${VM_USER}@${HOST} sudo rm -r "~/conferencev2/logs"

	ssh -i ${KEY_PATH} ${VM_USER}@${HOST} "cd conferencev2 && sudo make getLogs"
	
	scp -i ${KEY_PATH} -r ${VM_USER}@${HOST}:~/conferencev2/logs .
	ssh -i ${KEY_PATH} ${VM_USER}@${HOST} sudo rm -r "~/conferencev2/logs"

# make
# docker
# docker nginx
# docker envoy
# docker redis
# need to run on server!

serverDownloadLibs:
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

serverCreateStructure:
	ssh -i ${SSH_KEY_PATH} ${VM_USER}@${HOST} mkdir -p /conferencev2/html
	scp -i ${SSH_KEY_PATH} -r ssl ${VM_USER}@${HOST}:~/conferencev2/html

serverSetup: serverDownloadLibs serverCreateStructure