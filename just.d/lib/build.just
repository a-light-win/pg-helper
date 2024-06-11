import "./build/sqlc.just"
import "./build/proto.just"

# TODO: support multiple architectures, e.g. amd64, arm64, etc.
[no-cd,private]
_build_all release_mode="release": (_build release_mode 'amd64') (_build release_mode 'arm64')

[no-cd,private]
_build release_mode arch:_sqlc _generate_protos
  #!/usr/bin/env bash
  arch="{{ if arch == '' { arch() } else { arch } }}"
  if [ "$arch" == "x86_64" ]; then
    arch="amd64"
  elif [ "$arch" == "aarch64" ]; then
    arch="arm64"
  fi
  
  export CGO_ENABLED=0
  export GOARCH=${arch}
  debug_flag="{{ if release_mode == 'release' { '-s -w' } else { '' } }}"
  bin_name="{{ if release_mode == 'release' { 'pg-helper' } else { 'pg-helper-debug' } }}"

  echo "Building ${arch}/${bin_name} ..."

  version=$(git describe --tags --dirty 2>/dev/null)
  if [ $? -ne 0 ]; then
    version=v0-$(git describe --tags --always --dirty)
  fi
  go_version=$(go version|sed 's/go version go\(.*\)/\1/g')

  mkdir -p "dist/${arch}"
  go build -o "dist/${arch}/${bin_name}" -trimpath -ldflags "${debug_flag} -X main.Version=${version} -X 'main.GoVersion=${go_version}'" cmd/pg-helper/*.go || exit $?

  echo "Build ${arch}/${bin_name} ${version} success"

[no-cd,private]
_clean_build: _clean_sqlc _clean_protos
  rm -rf dist/
