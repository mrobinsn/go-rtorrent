#!/bin/bash

docker rm -f rutorrent
rm -rf tmp
mkdir tmp
docker run -d --name=rutorrent -v $(pwd)/tmp/data:/config -v $(pwd)/tmp/downloads:/downloads -e PGID=1000 -e PUID=1000 -p 80:80 -p 5000:5000 -p 51413:51413 -p 6881:6881/udp linuxserver/rutorrent
sleep 5
go test -v -race ./...

