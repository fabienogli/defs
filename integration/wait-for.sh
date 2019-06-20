#!/bin/sh
ADDRESS='docker:8080/file'

until [[ $(curl --write-out %{http_code} $ADDRESS) == "405" ]]; do
    printf '.'
    sleep 5
done