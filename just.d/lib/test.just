import "./build.just"

[no-cd,private]
_test_all: _build_all
  go test -coverprofile=coverage.out ./...

[no-cd,private]
_test_coverage:
  go tool cover -html=coverage.out

[no-cd,private]
_clean_coverage:
  rm -f coverage.out
