FROM golang:1.12

WORKDIR /go/src/loadbalancer/
COPY . .
ENV GO111MODULE on

#For HotReloading 
RUN ["go", "get", "github.com/githubnemo/CompileDaemon"]

#@todo Add multistage for prod with FROM build etc...
ENTRYPOINT CompileDaemon -log-prefix=false -build="go install ./..." -command="/go/bin/loadbalancer"