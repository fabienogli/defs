#!/bin/sh
FILENAME='test'
CONTENT='This is a test'
echo  $CONTENT > $FILENAME
sha256sum $FILENAME > "$FILENAME.sha256sum"
RESULT=$(curl -X POST -F filename=hello.txt -F file=@$FILENAME localhost:8080/file)
echo $RESULT