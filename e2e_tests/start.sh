#!/bin/bash

if [ $(docker compose ps --filter status=running |grep -v venom |wc -l) -lt 3 ]
then
  docker-compose up --build -d
fi

while [ $(docker compose ps --filter status=running |grep -v venom |wc -l) -lt 3 ]
do
  sleep 5
done

# Let time to smocker container to insert the mocks
sleep 30
