#!/bin/bash
docker build . -t panopticon
docker run  --rm panopticon
