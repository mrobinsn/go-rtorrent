#!/bin/bash

docker rm -f rutorrent
rm -rf tmp
mkdir tmp
docker run -d --name=rutorrent -p 8080:8080 -p 8000:8000 crazymax/rtorrent-rutorrent:latest
sleep 60
go test -v -race ./...

