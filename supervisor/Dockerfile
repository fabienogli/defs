FROM golang:1.12

WORKDIR /go/src/supervisor/
COPY . .

ENV GO111MODULE on

#For HotReloading 
RUN ["go", "get", "github.com/githubnemo/CompileDaemon"]

#@todo Add multistage for prod with FROM build etc...