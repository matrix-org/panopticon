# panopticon
panopticon records usage metrics from homeservers

## Building
To build panopticon, you need a working go installation.
To install dependencies, run:

```sh
go get github.com/mattn/go-sqlite3
go get github.com/go-sql-driver/mysql
```

To build, run:

```sh
go build
```

## Testing
There is a second `Dockerfile-testing` that builds panopticon to run the tests as above, as we probably want locally.

This only requires docker on your local workstation, no go install or dependencies required.

```sh
docker-tests.sh
```
To add new tests, crib exiting files in the `tests` directory.

# Deployment using docker image

Set the environment variables
 * `PANOPTICON_DB_DRIVER` (eg, mysql or sqlite) 
 * `PANOPTICON_DB` (go mysql connection string or filename for sqlite)
 * `PANOPTICON_PORT` (http port to expose panopticon on)

