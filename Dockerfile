FROM arangodb/arangodb:3.10.8 AS arango

FROM golang:1.20.5-bookworm AS devel

COPY --from=arango /usr/bin/arangodump /usr/bin
COPY --from=arango /etc/arangodb3/arangodump.conf /etc/arangodb3/arangodump.conf

ENV APP_PATH /go/src/devel

WORKDIR ${APP_PATH}

COPY go.mod go.sum $APP_PATH/
RUN go mod download

RUN go mod download

COPY . $APP_PATH

RUN make build