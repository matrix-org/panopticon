#!/bin/bash
docker build . -t panopticon-tests -f Dockerfile-tests
docker run -t --rm panopticon-tests
