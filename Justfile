
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
  go build -o dist/pg-helper -ldflags "-X main.Version=${version}" cmd/pg-helper/*.go || exit $?
  echo "Build pg-helper ${version} success"

strip: build
  #!/usr/bin/env bash
  cd dist/
  objcopy --only-keep-debug pg-helper pg-helper.dbg
  objcopy --strip-unneeded pg-helper
  objcopy --add-gnu-debuglink=pg-helper.dbg pg-helper


clean:
  rm -rf dist/
  rm -rf internal/db/

serve: strip
  {{ env('DOCKER_CMD', 'podman')}} compose up --force-recreate --build
  
