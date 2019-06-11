FROM golang:1.12 as build
RUN groupadd -g 2000 go \
 && useradd -m  -u 2001 -g go go
WORKDIR /go/src/storage/

COPY go.mod .
COPY go.sum .

ENV GO111MODULE on

RUN go mod download
#For HotReloading
RUN ["go", "get", "github.com/githubnemo/CompileDaemon"]
RUN apt-get update && apt-get install at
RUN /etc/init.d/atd start


FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/storage /go/bin/storage

USER appuser

ENTRYPOINT [ "/go/bin/hello" ]