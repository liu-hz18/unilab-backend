# Unilab-backend

### Modify Configs

see `./conf.ini`

### Local Prepare

```
mysql
golang
```
### Local Develop Run

```
make dev-build-local
make dev-run-local
```

### Local Release Run

```
make build-local
make run-local
```

### Deploy using Docker

First please modify `APP_MOUNT_DIR` and `MYSQL_MOUNT_DIR` in `Makefile` according to your file system.

```
make mysql  # if needed
make build-docker
make run-docker  # in develop mode, run `make dev-run-docker`
```

