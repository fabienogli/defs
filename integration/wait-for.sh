#!/bin/sh
ADDRESS='docker:8080/file'

until [[ $(curl --output /dev/null --silent --write-out %{http_code} $ADDRESS) == "405" ]]; do
    printf '.'
    sleep 5
done

echo "I am tired so I'll take a 5 seconds nap"
sleep 5
echo "Now it should be good"