FROM golang:1.12 as build
RUN groupadd -g 2000 go \
 && useradd -m  -u 2001 -g go go
WORKDIR /go/src/storage/

COPY go.mod .

#Uncomment this when having dependencies
#COPY go.sum . 

ENV GO111MODULE on

RUN go mod download
#For HotReloading
RUN ["go", "get", "github.com/githubnemo/CompileDaemon"]

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/storage /go/bin/storage

USER appuser

ENTRYPOINT [ "/go/bin/hello" ]