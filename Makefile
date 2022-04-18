DB_USER=root
DB_PASSWORD=123456
DB_HOST=127.0.0.1
DB_PORT=3307

APP_MOUNT_DIR=d/unilab-backend-mount
APP_INNER_DIR=/unilab-files

MYSQL_MOUNT_DIR=d/mysql-docker
MYSQL_INNER_DIR=/var/lib/mysql

EXPOSE_PORT=1323

.PHONY: mysql build-db build-table drop-db check lint clean dev-build-local dev-run-local build-local run-local build-docker dev-run-docker run-docker

build-db:
	mysql -h $(DB_HOST) -P $(DB_PORT) -u $(DB_USER) -p$(DB_PASSWORD) < ./mysql/create_db.sql 

build-table:
	mysql -h $(DB_HOST) -P $(DB_PORT) -u $(DB_USER) -p$(DB_PASSWORD) < ./mysql/create_table.sql

drop-db:
	mysql -h $(DB_HOST) -P $(DB_PORT) -u $(DB_USER) -p$(DB_PASSWORD) < ./mysql/drop_db.sql

check:
	go tool vet . |& grep -v vendor; true
	gofmt -w .

lint:
	golint ./...

clean:
	go clean -i .

# local (develop)
dev-build-local: drop-db build-db build-table

dev-run-local: 
	go run main.go

# local (release)
build-local: build-db build-table
	go build -v -o main .

run-local: 
	./main

# docker deploy
mysql:
	docker pull mysql:latest
	docker run --name mysql -p $(DB_PORT):3306 -e MYSQL_ROOT_PASSWORD=$(DB_PASSWORD) -v $(MYSQL_MOUNT_DIR):$(MYSQL_INNER_DIR) -d mysql

build-docker:
	docker build -t unilab-backend-docker .

dev-run-docker: drop-db build-db build-table
	docker run --name unilab-backend --link mysql:mysql -p $(EXPOSE_PORT):$(EXPOSE_PORT) -v $(APP_MOUNT_DIR):$(APP_INNER_DIR) -d unilab-backend-docker 

run-docker: build-db build-table
	docker run --name unilab-backend --link mysql:mysql -p $(EXPOSE_PORT):$(EXPOSE_PORT) -v $(APP_MOUNT_DIR):$(APP_INNER_DIR) -d unilab-backend-docker 
