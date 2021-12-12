include .env
export $(shell sed 's/=.*//' .env)

prep-linux:
	curl -L https://github.com/protocolbuffers/protobuf/releases/download/v3.19.1/protoc-3.19.1-linux-x86_64.zip --output /tmp/protoc.zip
	sudo unzip -o /tmp/protoc.zip -d /usr/local bin/protoc
	sudo unzip -o /tmp/protoc.zip -d /usr/local bin/protoc
	sudo unzip -o /tmp/protoc.zip -d /usr/local 'include/*'
	sudo chmod a+x /usr/local/bin/protoc
	rm -rf /tmp/protoc.zip
	go get -u google.golang.org/protobuf/{proto,protoc-gen-go} || true
	sudo sysctl -w vm.max_map_count=262144

gen:
	export GOPATH=$$HOME/go
	export PATH=$$PATH:/$$GOPATH/bin
	protoc -I proto/ proto/*.proto --go_out=plugins=grpc:proto
	mv proto/*.pb.go .

dev:
	docker kill $$(docker ps -q) || true
	docker network rm ${PROJECT_NAME}_default; true
	docker network create -d bridge ${PROJECT_NAME}_default ; true
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate --remove-orphans testnats
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate --remove-orphans testpsql
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate --remove-orphans testes01
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate --remove-orphans testes02
	docker run --network ${PROJECT_NAME}_default willwill/wait-for-it testes01:9200 -- echo "elastic is up"
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate --remove-orphans testkib
	docker run --network ${PROJECT_NAME}_default willwill/wait-for-it testpsql:5432 -- echo "database is up"
	docker-compose -p ${PROJECT_NAME} run testpsql dropdb -h testpsql -U ${DB_USER} -w ${DB_NAME} || true
	docker-compose -p ${PROJECT_NAME} run testpsql createdb -h testpsql -U ${DB_USER} -w ${DB_NAME}

es_analyze:
	curl -XPOST http://localhost:9200/article/_analyze

db:
	docker exec -ti ${PROJECT_NAME}_testpsql_1 psql -U testpsql -d testpsql

vet:
	go vet ./...

test:
	go test -v ./... -coverprofile=coverage.html

cov:
	go tool cover -html=coverage.html

dockerbuild:
	docker build -t dddcqrs/http -f Dockerfile.http .
	docker build -t dddcqrs/service -f Dockerfile.service .
	docker build -t dddcqrs/command -f Dockerfile.command .
	docker build -t dddcqrs/query -f Dockerfile.query .
	docker build -t dddcqrs/eventstore -f Dockerfile.eventstore .

dockertest:
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate --remove-orphans testcommand
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate --remove-orphans testquery
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate --remove-orphans testservice
	docker build -t dddcqrs/http -f Dockerfile.http .
	docker run -ti \
	--network ${PROJECT_NAME}_default \
	-e PROJECT_NAME=${PROJECT_NAME} \
	-e APP_NAME=${APP_NAME} \
	-e DB_HOST='testpsql' \
	-e DB_USER=${DB_USER} \
	-e DB_PASS=${DB_PASS} \
	-e DB_NAME=${DB_NAME} \
	-e NATS_URL='nats://testnats:4222' \
	-e GRPC_ADDRESS='testservice:4040' \
	-e ELASTIC_ADDRESS='testes01:9200' \
	${PROJECT_NAME}/http /svc/scripts/test.sh

post:
	curl -X DELETE 'http://localhost:9200/article'
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"robohnya surau kami","body":"tentang orang yang lupa hablumminannas","author":"aa navis"}'
	sleep 1.1
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"radical candor","body":"being honest is good","author":"kim scott"}'
	sleep 1.1
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"metamorfosis","body":"the longest short story of kafka","author":"franz kafka"}'
	sleep 1.1
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"the rosie project","body":"back at the bar","author":"graeme simsion"}'
	sleep 1.1
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"moby dick","body":"epic saga of one legend fanatic","author":"herman melville"}'
	sleep 1.1
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"ibunda","body":"merupakan sosok perempuan yang hidup di masa revolusi demokratik rusia","author":"maxim gorki"}'
	sleep 1.1
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"the name of the rose","body":"imagine a medieval castle","author":"umberto eco"}'
	sleep 1.1
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"scandal","body":"when suguro and kurimoto arrived","author":"endo"}'
	sleep 1.1
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"the alchemist","body":"the boy s name was santiago","author":"paulo coelho"}'
	sleep 1.1
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"momo","body":"kisah momo berlangsung di negeri hayalan","author":"michael ende"}'
	sleep 1.1
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"3 years","body":"jatuh cinta ke pada yulia, alexei tak mau lama-lama menunggu melamarnya","author":"anton chekov"}'
	sleep 1.1
post-2:
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"pygmy","body":"nicknamed pygmy for his diminutive size","author":"chuck palahniuk"}'

get-list:
	curl -X GET 'http://localhost:8000/api/articles'

get-page:
	curl -X GET 'http://localhost:8000/api/articles?page=1&limit=5'

get-query:
	curl -X GET 'http://localhost:8000/api/articles?page=1&limit=5&query=surau'

get-author:
	curl -X GET 'http://localhost:8000/api/articles?page=1&limit=5&author=anton'

get-item:
	curl -X GET 'http://localhost:8000/api/article/71'
