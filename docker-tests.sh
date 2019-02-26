#!/bin/bash
docker build . -t panopticon
docker run -t --rm panopticon
