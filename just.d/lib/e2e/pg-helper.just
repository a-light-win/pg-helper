import "../build.just"

[no-cd,private]
_copy_pg_helper_to_e2e:
  #!/usr/bin/env bash
  if [ ! -x dist/amd64/pg-helper ]; then
    echo >&2 "pg-helper not built yet"
    exit 1
  fi
  rsync -av dist/amd64/ e2e/dist/

[no-cd,private]
_clean_pg_helper_in_e2e:
  rm -rf e2e/dist
