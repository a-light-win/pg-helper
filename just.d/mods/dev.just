
import "../lib/vars.just"

builder_version := '0.1.12'
builder_cmd := container_provider + " run -it --rm -v `pwd`:/pg-helper -v /tmp/go:/go ghcr.io/a-light-win/pg-helper/bulder:" + builder_version
just_cmd := 'just --unstable'

# Build pg-helper inside a container
[no-cd]
build: go-cache
  {{ builder_cmd }} {{ just_cmd }} build

# Clean the build files
[no-cd]
clean: go-cache
  {{ builder_cmd }} {{ just_cmd }} build clean

# Runing the unit tests inside a container
[no-cd]
test: go-cache
  {{ builder_cmd }} {{ just_cmd }} build test

# Clean the test results
[no-cd]
clean-test: go-cache
  {{ builder_cmd }} {{ just_cmd }} build clean-test

# Create new migration
[no-cd]
new-migration name: go-cache
  {{ builder_cmd }} {{ just_cmd }} build new-migration {{ name }}
  

# Enter the container shell for debugging or testing
[no-cd]
bash: go-cache
  {{ builder_cmd }} bash

[no-cd, private]
go-cache:
  mkdir -p /tmp/go
