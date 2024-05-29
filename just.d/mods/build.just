
import "../lib/build.just"

[no-cd,private]
build: _build_all

[no-cd]
clean: _clean_build
  rm -rf dist/pg-helper
