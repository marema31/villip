#!/bin/sh

/opt/smocker &
pid=$!

curl -X POST http://localhost:8081/reset

for mock in /mocks/*.yaml /mocks/*.yml
do
    mockname=`basename $mock|sed -e 's/\..*//'`
    if [ "$mockname" != "*" ]
    then
        echo "inserting $mockname"
        curl -X POST -H "Content-Type: application/x-yaml" --data-binary "@-" http://localhost:8081/mocks < $mock
    fi
done

wait $pid