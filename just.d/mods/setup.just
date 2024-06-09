
import "../lib/setup.just"

[no-cd]
container-env: _update_repo_cache && \
  _setup_podman

[no-cd]
build-env: _update_repo_cache && \
  (_setup 'file') \
  _setup_proto \
  _setup_sqlc \
  _setup_goose

[no-cd]
builder version: _setup_podman
  #!/usr/bin/env bash
  {{ container_provider }} build -t ghcr.io/a-light-win/pg-helper/bulder:{{ version }} .
  {{ container_provider }} tag ghcr.io/a-light-win/pg-helper/bulder:{{ version }} ghcr.io/a-light-win/pg-helper/bulder:latest
