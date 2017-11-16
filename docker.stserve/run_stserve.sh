#!/bin/bash
set -x
docker run -d \
        --name stserve \
        --restart always \
        --user "$(id -u):$(id -g)" \
        --volumes-from "syncthing" \
        -p 8765:8765 \
        mdevey/stserve "$@"
timeout 10s docker logs -f stserve || true
