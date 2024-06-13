
import "../lib/vars.just"

builder_version := '0.1.15'
go_path := env('GOPATH', "$HOME/go")
go_cache := env('GOCACHE', '$HOME/.cache/go-build')
builder_cmd := container_provider + " run -it --rm -v `pwd`:/pg-helper -v " + go_path + ":/go -v " + go_cache + ":/go/cache ghcr.io/a-light-win/pg-helper/builder:" + builder_version
just_cmd := 'just --unstable'

# Build pg-helper inside a container
[no-cd]
build-all release_mode='release': go-cache
  {{ builder_cmd }} {{ just_cmd }} build build-all "{{ release_mode }}"

[no-cd]
build release_mode='release' arch='': go-cache
  {{ builder_cmd }} {{ just_cmd }} build build "{{ release_mode }}" "{{ arch}}"

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
  mkdir -p "{{ go_path }}"
  mkdir -p "{{ go_cache }}"
