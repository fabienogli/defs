FROM golang:1.12

WORKDIR /go/src/app/
COPY . .

#For HotReloading 
RUN ["go", "get", "github.com/githubnemo/CompileDaemon"]
ENTRYPOINT CompileDaemon -log-prefix=false -build="go install ./..." -command="/go/bin/app"

#@todo Add multistage for prod with FROM build etc...