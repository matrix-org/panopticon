from golang

RUN apt-get update && apt-get install sqlite3 && apt-get clean
WORKDIR /go/src/panopticon
RUN go get github.com/mattn/go-sqlite3
RUN go get github.com/go-sql-driver/mysql

COPY ./runtests.sh /go/src/panopticon
COPY ./tests /go/src/panopticon/tests
COPY ./main.go /go/src/panopticon
RUN go build
CMD ./runtests.sh

