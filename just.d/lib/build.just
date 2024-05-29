import "./build/sqlc.just"
import "./build/proto.just"
import "./build/strip.just"

# TODO: support multiple architectures, e.g. amd64, arm64, etc.
[no-cd,private]
_build_all: _sqlc _generate_protos && (_strip 'pg-helper')
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

[no-cd,private]
_clean_build: _clean_sqlc _clean_protos && (_clean_strip 'pg-helper')
  rm -rf dist/pg-helper
