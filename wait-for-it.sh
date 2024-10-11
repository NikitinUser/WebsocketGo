#!/bin/bash

if [ -z "$1" ]; then
    echo "Использование: $0 <URL>"
    exit 1
fi

URL="http://$1"

while true; do
    HTTP_CODE=$(curl --write-out "%{http_code}" --silent --output /dev/null "$URL")

    if [ "$HTTP_CODE" -eq 200 ]; then
        echo "Сайт доступен: $URL"
        break
    fi
    sleep 5
done

./main
