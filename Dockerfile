FROM golang:1.24 AS build_img
ENV APP_DIR=/app
RUN mkdir -p $APP_DIR
COPY *.go $APP_DIR
WORKDIR $APP_DIR
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -gcflags "all=-N -l" -o /villip

ENTRYPOINT /villip

FROM scratch

COPY --from=build_img /villip /usr/bin/villip

ENTRYPOINT ["/usr/bin/villip" ]
