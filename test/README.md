# Test

## Unit tests

Before running the unit tests starts postgreSQL in detached mode:

```shell
docker run -d -e POSTGRES_USER=admin -e POSTGRES_PASSWORD=admin -p 5432:5432 --name postgresql postgres
```

Then run the unit tests:

```shell
go test ./...
```

