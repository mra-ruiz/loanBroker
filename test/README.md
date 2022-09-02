# Test

## Unit tests

Before running the unit tests starts postgreSQL in detached mode:

```shell
docker run -d -e POSTGRES_USER=admin -e POSTGRES_PASSWORD=admin -p 5433:5432 docker.io/postgres:latest
```

Then run the unit tests:

```shell
go test ./...
```