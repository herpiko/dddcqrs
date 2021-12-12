include .env
export $(shell sed 's/=.*//' .env)

gen-prep:
	curl -L https://github.com/protocolbuffers/protobuf/releases/download/v3.19.1/protoc-3.19.1-linux-x86_64.zip --output /tmp/protoc.zip
	sudo unzip -o /tmp/protoc.zip -d /usr/local bin/protoc
	sudo unzip -o /tmp/protoc.zip -d /usr/local bin/protoc
	sudo unzip -o /tmp/protoc.zip -d /usr/local 'include/*'
	sudo chmod a+x /usr/local/bin/protoc
	rm -rf /tmp/protoc.zip
	go get -u google.golang.org/protobuf/{proto,protoc-gen-go} || true

gen:
	export GOPATH=$$HOME/go
	export PATH=$$PATH:/$$GOPATH/bin
	protoc -I proto/ proto/*.proto --go_out=plugins=grpc:proto
	mv proto/*.pb.go .

dev:
	docker kill $$(docker ps -q) || true
	sudo rm -rf /tmp/data || true
	mkdir -p data/psql data/elastic1 data/elastic2
	docker network rm ${PROJECT_NAME}_default; true
	docker network create -d bridge ${PROJECT_NAME}_default ; true
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate testnats
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate testpsql
	sudo sysctl -w vm.max_map_count=262144
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate testes01
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate testes02
	#docker-compose -p ${PROJECT_NAME} up -d --force-recreate testes03
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate testkib01
	docker run --network ${PROJECT_NAME}_default willwill/wait-for-it testpsql:5432 -- echo "database is up"
	docker-compose -p ${PROJECT_NAME} run testpsql createdb -h testpsql -U ${DB_USER} -w ${DB_NAME}

db:
	docker exec -ti ${PROJECT_NAME}_testpsql_1 psql -U testpsql -d testpsql

run:
	go run cmd/main.go

vet:
	go vet ./...

test:
	go test ./... -coverprofile=coverage.html

cov:
	go tool cover -html=coverage.html

dockerbuild:
	docker build -t ${PROJECT_NAME}-api .

dockertest:
	docker run -ti \
	--network ${PROJECT_NAME}_default \
	-e PROJECT_NAME=${PROJECT_NAME} \
	-e APP_NAME=${APP_NAME} \
	-e DB_HOST='testpsql' \
	-e DB_USER=${DB_USER} \
	-e DB_PASS=${DB_PASS} \
	-e DB_NAME=${DB_NAME} \
	${PROJECT_NAME}-api /app/scripts/test.sh

post:
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"title1","body":"body1","author":"author1"}'
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"robohnya surau kami","body":"loremipsum1","author":"aa navis"}'
	curl -X POST http://localhost:8000/api/articles -H 'Content-Type: application/json' -d '{"title":"radical candor","body":"loremipsum2","author":"kim scott"}'
get:
	curl -X GET http://localhost:8000/api/articles

get-page:
	curl -X GET 'http://localhost:8000/api/articles?page=1&limit=5'

get-search:
	curl -X GET 'http://localhost:8000/api/articles?page=1&limit=5&search=roboh'
