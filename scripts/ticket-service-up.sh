#!/bin/bash

export JWT_KEY=jwtkey 
export STAFF_TOKEN=stafftoken 
export DB_USER=dbuser 
export DB_PASSWORD=dbpassword 
export DB_NAME=dbname 

docker-compose -f docker-compose.yaml up -d