DB_USER=root
DB_PASSWORD=123456
DB_HOST=127.0.0.1
DB_PORT=3306

.PHONY: build_db

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

build:
	go build -v -o main.exe .

dev-run: drop-db build-db build-table
	go run main.go

run: build-db build-table build
	./main.exe
