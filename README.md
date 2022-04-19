# Unilab-backend

### Modify Configs

see `./conf.ini`

### Requirements

```
mysql
golang
docker
docker-compose
```
### Local Develop Run

```
make build-db # run `make rebuild-db` if you want to clear db
make dev-run-local
```

### Local Release Run

```
make build-db # run `make rebuild-db` if you want to clear db
make run-local
```

### Deploy using Docker

First please modify `APP_MOUNT_DIR` and `MYSQL_MOUNT_DIR` in `Makefile` according to your file system.

```
make build-docker
make run-docker
# run `make rebuild-db` if you want to clear db
```

then type `localhost/api` in your browser to access our backend. 

#### Stop services

```
make stop-docker
```

