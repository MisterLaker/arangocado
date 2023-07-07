FROM arangodb/arangodb:3.10.8 AS arango

FROM golang:1.20.5-bookworm AS devel

COPY --from=arango /usr/bin/arangodump /usr/bin
COPY --from=arango /usr/bin/arangorestore /usr/bin
COPY --from=arango /etc/arangodb3/arangodump.conf /etc/arangodb3/arangodump.conf
COPY --from=arango /etc/arangodb3/arangorestore.conf /etc/arangodb3/arangorestore.conf

ENV APP_PATH /opt/arangocado

ENV LINTER_VERSION v1.53.3

RUN set -ex \
    && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin $LINTER_VERSION

WORKDIR ${APP_PATH}

COPY go.mod go.sum $APP_PATH/
RUN go mod download

COPY . $APP_PATH

RUN make build

FROM alpine:3.18

RUN apk --no-cache add ca-certificates libc6-compat

WORKDIR /opt/arangocado

COPY --from=arango /usr/bin/arangodump /usr/bin
COPY --from=arango /usr/bin/arangorestore /usr/bin
COPY --from=arango /etc/arangodb3/arangodump.conf /etc/arangodb3/arangodump.conf
COPY --from=arango /etc/arangodb3/arangorestore.conf /etc/arangodb3/arangorestore.conf

COPY --from=devel /opt/arangocado/build/* /opt/arangocado/
COPY --from=devel /opt/arangocado/config.yaml /opt/arangocado/config.yaml
