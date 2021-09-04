#!/bin/bash
# shellcheck disable=SC2046

if [ $(docker compose ps --filter status=running |grep -vc venom) -lt 3 ]
then
  docker-compose up --build -d
fi

while [ $(docker compose ps --filter status=running |grep -vc venom) -lt 3 ]
do
  sleep 5
done

# Let time to smocker container to insert the mocks
sleep 30
