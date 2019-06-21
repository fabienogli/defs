#!/bin/sh

/etc/init.d/atd start
CompileDaemon -log-prefix=false -build="go install ./..." -command="/go/bin/storage"
