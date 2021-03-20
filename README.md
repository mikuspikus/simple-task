# simple-task

For running all test for internal/api
```
cd internal/api
make test
```
`make` probably require `sudo` or `su` because of _docker-compose_ usage. internal/api/server_test.go test HTTP server with mock db. internal/api/db_test.go test PSQLdb from db.go (main reason for _docker-compose_ usage).
