
_install-goose:
  #!/usr/bin/env bash
  which goose &> /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing goose ..."
    GOBIN=~/.local/bin go install github.com/pressly/goose/v3/cmd/goose@latest
  fi

new-migrate name: _install-goose
  goose -dir db/migrations create {{ name }} sql

_sqlc:
  {{ env('DOCKER_CMD', 'podman') }} run -it --rm -v `pwd`:`pwd` -w `pwd` docker.io/sqlc/sqlc generate 

_install-protoc:
  #!/usr/bin/env bash
  which protoc &> /dev/null
  if [ $? -eq 0 ]; then
    exit 0
  fi

  echo "Installing protoc ..."
  if which apt-get &> /dev/null ; then
    sudo apt install -y protobuf-compiler
  elif which pacman &> /dev/null ; then
    sudo pacman -S protobuf
  fi 

_install-protoc-gen-go: _install-protoc
  #!/usr/bin/env bash
  which protoc-gen-go &> /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing protoc-gen-go ..."
    GOBIN=~/.local/bin go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  fi

_install-protoc-gen-go-grpc: _install-protoc-gen-go
  #!/usr/bin/env bash
  which protoc-gen-go-grpc &> /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing protoc-gen-go-grpc ..."
    GOBIN=~/.local/bin go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  fi

_generate-protos: _install-protoc-gen-go-grpc
  #!/usr/bin/env bash
  echo "Generating protos ..."
  protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/proto/*.proto

build: _sqlc _generate-protos && _strip
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

serve: build
  {{ env('DOCKER_CMD', 'podman')}} compose up --force-recreate --build
  
_clean-protos:
  rm -rf api/proto/*.pb.go

_clean-sqlc:
  rm -rf internal/db/db.go internal/db/models.go internal/db/*.sql.go

clean: _clean-sqlc _clean-protos
  rm -rf dist/

