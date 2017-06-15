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
To run tests, run `./runtests.sh`
To add new tests, crib exiting files in the `tests` directory.
