#!/bin/bash
# Set the path to your database
echo "Building docker image"
sudo docker build -t llforum .
echo ""
echo "Pruning docker images"
sudo docker system prune -f
echo ""
echo "Showing docker images"
sudo docker images
echo "Running docker container"
sudo docker run -p 8080:8080 --name llforum -v "$(pwd)/data:/data" -it --rm llforum
# We set an env var for easier db access
