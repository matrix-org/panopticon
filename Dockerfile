FROM golang:1.18

RUN apt-get -yq update && apt-get -yq install sqlite3 && apt-get -yq clean
WORKDIR /go/src/panopticon

COPY ./runtests.sh /go/src/panopticon
COPY ./tests /go/src/panopticon/tests
COPY ./*.go /go/src/panopticon
COPY ./go.mod /go/src/panopticon
COPY ./go.sum /go/src/panopticon
RUN go build
RUN ./runtests.sh

FROM debian

WORKDIR /root/
COPY --from=0 /go/src/panopticon/panopticon .
COPY ./docker-start.sh .
CMD ["/root/docker-start.sh"]

