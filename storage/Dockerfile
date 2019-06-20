FROM golang:1.12 as build

WORKDIR /go/src/storage/
RUN useradd --user-group go
COPY go.mod .
COPY go.sum .
ENV GO111MODULE on
RUN go mod download
#For HotReloading
RUN go get github.com/githubnemo/CompileDaemon
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go install ./...
RUN apt-get update && apt-get install at -y
RUN /etc/init.d/atd start

FROM scratch as prod

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build --chown=go /go/bin/storage /go/bin/storage

USER go