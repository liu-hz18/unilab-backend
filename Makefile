DB_USER=root
DB_PASSWORD=123456
DB_HOST=127.0.0.1
DB_PORT=3307

.PHONY: mysql build-db build-table drop-db check lint clean dev-build-local dev-run-local build-local run-local build-docker dev-run-docker run-docker

build-db:
	mysql -h $(DB_HOST) -P $(DB_PORT) -u $(DB_USER) -p$(DB_PASSWORD) < ./mysql/create_db.sql 
	mysql -h $(DB_HOST) -P $(DB_PORT) -u $(DB_USER) -p$(DB_PASSWORD) < ./mysql/create_table.sql

drop-db:
	mysql -h $(DB_HOST) -P $(DB_PORT) -u $(DB_USER) -p$(DB_PASSWORD) < ./mysql/drop_db.sql

rebuild-db: drop-db build-db

check:
	go tool vet . |& grep -v vendor; true
	gofmt -w .

lint:
	~/go/bin/golint ./...

clean:
	go clean -i .

# local (develop)
dev-run-local: 
	go run main.go

# local (release)
run-local:
	go build -v -o main .
	./main

# docker deploy
build-docker:
	docker-compose build

run-docker:
	docker-compose up -d

stop-docker:
	docker-compose stop 
# docker-compose kill

# sudo docker run --name mysql -p 3307:3306 -e MYSQL_ROOT_PASSWORD=123456 -v ~/mysql-docker:/var/lib/mysql -d mysql
# sudo docker build -t unilab-backend-docker .
# sudo docker run --name unilab-backend --link mysql:mysql -p 8080:1323 -v ~/unilab-backend-mount/runtime:/unilab-files/runtime -v ~/unilab-backend-mount/upload:/unilab-files/upload -d unilab-backend-docker

