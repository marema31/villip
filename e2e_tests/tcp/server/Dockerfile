FROM golang:1.13 AS build_img
ENV APP_DIR=/app
RUN mkdir -p $APP_DIR
COPY *.go $APP_DIR
WORKDIR $APP_DIR

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -gcflags "all=-N -l" -o /tcp_server

ENTRYPOINT /tcp_server

FROM scratch

COPY --from=build_img /tcp_server /usr/bin/tcp_server

ENTRYPOINT ["/usr/bin/tcp_server"]
