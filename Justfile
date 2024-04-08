
_install-goose:
  #!/usr/bin/env bash
  which goose > /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing goose ..."
    GOBIN=~/.local/bin go install github.com/pressly/goose/v3/cmd/goose@latest
  fi

new-migrate name: _install-goose
  goose -dir db/migrations create {{ name }} sql

_sqlc:
  {{ env('DOCKER_CMD', 'podman') }} run -it --rm -v `pwd`:`pwd` -w `pwd` docker.io/sqlc/sqlc generate 

build: _sqlc && _strip
  #!/usr/bin/env bash
  echo "Building pg-helper ..."

  version=$(git describe --tags --dirty 2>/dev/null)
  if [ $? -ne 0 ]; then
    version=v0-$(git describe --tags --always --dirty)
  fi
  go_version=$(go version|sed 's/go version go\(.*\)/\1/g')
  go build -o dist/pg-helper -trimpath -ldflags "-X main.Version=${version} -X 'main.GoVersion=${go_version}'" cmd/pg-helper/*.go || exit $?

  echo "Build pg-helper ${version} success"

_strip:
  #!/usr/bin/env bash
  echo "Stripping pg-helper binary ..."

  cd dist/
  file pg-helper |grep -q "not stripped"
  if [ $? -ne 0 ]; then
    echo "pg-helper binary already stripped"
    exit 0
  fi

  objcopy --only-keep-debug pg-helper pg-helper.dbg
  objcopy --strip-unneeded pg-helper
  objcopy --add-gnu-debuglink=pg-helper.dbg pg-helper

  echo "Stripping pg-helper binary success"

clean:
  rm -rf dist/
  rm -rf internal/db/

serve: build
  {{ env('DOCKER_CMD', 'podman')}} compose up --force-recreate --build
  
