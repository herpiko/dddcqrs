include .env
export $(shell sed 's/=.*//' .env)

dev:
	docker network create -d bridge testscope_default ; true
	docker-compose -p testscope up -d --force-recreate testdb
	docker run --network testscope_default willwill/wait-for-it testdb:5432 -- echo "database is up"
	docker-compose -p testscope run testdb createdb -h testdb -U ${DB_USER} -w ${DB_NAME} || true


run:
	go run .
