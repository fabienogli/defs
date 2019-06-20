FROM golang:1.12 as build

WORKDIR /go/src/supervisor/
RUN useradd --user-group go
COPY go.mod .
COPY go.sum .
ENV GO111MODULE on
RUN go mod download
#For HotReloading
RUN go get github.com/githubnemo/CompileDaemon
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go install ./...

FROM scratch as prod

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build --chown=go /go/bin/supervisor /go/bin/supervisor

USER go