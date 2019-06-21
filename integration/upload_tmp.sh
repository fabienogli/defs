#!/bin/sh

ADDRESS='docker:8080/file'

## UPLOADING FILE
FILENAME='test'
CONTENT='This is a test'
echo "Writing the file"
echo  $CONTENT > $FILENAME
OG_HASH=$(sha256sum $FILENAME | cut -d " " -f 1)

echo "Uploading the file"
RESULT=$(curl -X POST -F ttl="1 minute" -F filename=$FILENAME -F file=@$FILENAME --silent $ADDRESS)

rm $FILENAME

KEY=$(echo $RESULT | cut -d ":" -f 1 | tr -d \" | tr -d \{)
RETRIEVED_HASH=$(echo $RESULT | cut -d ":" -f 2 | tr -d \" | tr -d \})

if [[ $KEY != "hash" ]];then
    echo "Not the good key, here is the full response"
    echo $RESULT
    exit 1
fi

## DOWNLOADING FILE
OUTPUT_FILE="output"
curl --silent $ADDRESS/$RETRIEVED_HASH > $OUTPUT_FILE

NEW_HASH=$(sha256sum $OUTPUT_FILE | cut -d " " -f 1)

rm $OUTPUT_FILE

if [[ $NEW_HASH != $OG_HASH ]]; then
    echo "Different hash for same file"
    echo "Original Hash: $OG_HASH"
    echo "New hash: $NEW_HASH"
    exit 1
fi

# Waiting for the ttl to end
echo "Waiting for the ttl to end"
sleep 61
RESULT=$(curl --silent $ADDRESS/$RETRIEVED_HASH)
KEY=$(echo $RESULT | cut -d ":" -f 1 | tr -d \" | tr -d \{)
MSG=$(echo $RESULT | cut -d ":" -f 2 | tr -d \" | tr -d \})

if [[ $KEY != "error" ]]; then
    echo "there should be an error here instead have"
    echo $MSG
    exit 1
fi

exit 0