#!/bin/sh
set -e

# start
#   read the docker-compose.yml based project file
#   strip out anything not official docker-compose
#   replace the entrypoint of each service. if an existing entrypoint has been set, prepend this to command
#   override any user config to apply root
#   append volume mounts for the new entrypoint, build scripts, run scripts and preferred user id config
#   write the the new docker-compose file to a working hidden directory
#   run docker compose up -d on the project

# stop
#   check the working hidden directory exists. if not, refuse to run
#   run docker compose down

# destroy
#   check the working hidden directory exists. if not, refuse to run
#   run docker compose down -v

# rebuild
#   run destroy
#   run start

# restart
#   run stop
#   run start

# logs
#   run docker compose logs

# ssh
#   run docker compose exec sh
