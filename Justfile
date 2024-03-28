
sqlc:
  {{ env('DOCKER_CMD', 'podman') }} run -it --rm -v `pwd`:`pwd` -w `pwd` docker.io/sqlc/sqlc generate 

_new-migrate action name:
  {{ env('DOCKER_CMD', 'podman') }} run -v `pwd`:`pwd` -w `pwd` docker.io/migrate/migrate create -ext sql -dir db/migrations {{ action }}_{{ name }}

new-table name: (_new-migrate "create_table" name)

build: sqlc
  #!/usr/bin/env bash
  version=$(git describe --tags --dirty 2>/dev/null)
  if [ $? -ne 0 ]; then
    version=v0-$(git describe --tags --always --dirty)
  fi
  go build -o dist/pg-helper -ldflags "-X main.Version=${version}" cmd/pg-helper/*.go

clean:
  rm -rf dist/
  rm -rf internal/db/

serve: build
  ./dist/pg-helper serve --config=tests/config.yaml
  
