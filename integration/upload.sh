#!/bin/sh
## UPLOADING FILE
FILENAME='test'
CONTENT='This is a test'
echo "Writing the file"
echo  $CONTENT > $FILENAME
OG_HASH=$(sha256sum $FILENAME)

echo "Uploading the file"
RESULT=$(curl -X POST -F filename=$FILENAME -F file=@$FILENAME $ADDRESS)
KEY=$(echo $RESULT | cut -d ":" -f 1 | tr -d \" | tr -d \{)
RESPONSE=$(echo $RESULT | cut -d ":" -f 2 | tr -d \" | tr -d \})

echo "key is $KEY"
echo "response is $RESPONSE"
if [[ $KEY != "hash" ]];then
    echo "Not the good key, here is the full response"
    echo $RESULT
    exit 1
fi

## DOWNLOADING FILE
OUTPUT_FILE="output"
curl $ADDRESS/$RESPONSE > $OUTPUT_FILE
NEW_HASH=$(sha256sum $OUTPUT_FILE)
if [[ NEW_HASH != OG_HASH ]]; then
    echo "Different hash for same file, file is \n"
    echo $OUTPUT_FILE
    exit 1
fi
exit 0

