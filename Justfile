
mod secrets "just.d/mods/secrets.just"
mod migration "just.d/mods/migration.just"
mod db "just.d/mods/db.just"
mod e2e "just.d/mods/e2e.just"

serve_compose_args := "--profile pg-13 --profile pg-14 --profile authelia -f " + env('DOCKER_CMD', 'podman') + "-compose.yaml"

import "just.d/lib.just"
import "just.d/sqlc.just"
import "just.d/proto.just"

build: sqlc generate-protos && (strip 'pg-helper')
  #!/usr/bin/env bash
  echo "Building pg-helper ..."

  version=$(git describe --tags --dirty 2>/dev/null)
  if [ $? -ne 0 ]; then
    version=v0-$(git describe --tags --always --dirty)
  fi
  go_version=$(go version|sed 's/go version go\(.*\)/\1/g')

  export CGO_ENABLED=0
  go build -o dist/pg-helper -trimpath -ldflags "-X main.Version=${version} -X 'main.GoVersion=${go_version}'" cmd/pg-helper/*.go || exit $?

  echo "Build pg-helper ${version} success"

test: build && cover
  go test -coverprofile=coverage.out ./...

[private]
cover:
  go tool cover -html=coverage.out

serve: build
  #!/usr/bin/env bash
  export PG_HELPER_VERSION=$(./dist/pg-helper version|awk '{print $2}')
  {{ env('DOCKER_CMD', 'podman')}}-compose {{ serve_compose_args }} up --force-recreate --build
  {{ env('DOCKER_CMD', 'podman')}}-compose {{ serve_compose_args }} stop

clean-serve:
  #!/usr/bin/env bash
  set -x
  export PG_HELPER_VERSION=$(./dist/pg-helper version|awk '{print $2}')
  {{ env('DOCKER_CMD', 'podman')}}-compose {{ serve_compose_args }} down
  
clean: clean-sqlc clean-protos clean-serve
  rm -rf dist/
