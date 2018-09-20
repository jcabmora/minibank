#!/bin/bash
# Requires docker

set pipefail

# remove all the running  containers
docker rm -f `docker ps -aq`

# Create minibank network if it does not exist
docker network ls -f "driver=bridge" | grep ' minibanknet ' > /dev/null || docker network create minibanknet

docker run -d --name mysql -e MYSQL_ROOT_PASSWORD=hobbes -v `pwd`/scripts:/docker-entrypoint-initdb.d:ro -d --network minibanknet mariadb:latest
docker run -d --name minibank -p 80:8080 --network minibanknet minibank:latest
