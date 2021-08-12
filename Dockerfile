FROM golang:1.13

RUN apt-get update && apt-get install sqlite3 && apt-get clean
WORKDIR /go/src/panopticon
RUN go get github.com/mattn/go-sqlite3
RUN go get github.com/go-sql-driver/mysql

COPY ./runtests.sh /go/src/panopticon
COPY ./tests /go/src/panopticon/tests
COPY ./main.go /go/src/panopticon
RUN go build
RUN ./runtests.sh

FROM debian

WORKDIR /root/
COPY --from=0 /go/src/panopticon/panopticon .
COPY ./docker-start.sh .
CMD ["/root/docker-start.sh"]

