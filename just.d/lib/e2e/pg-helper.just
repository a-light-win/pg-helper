import "../build.just"

[no-cd,private]
_copy_pg_helper_to_e2e:
  #!/usr/bin/env bash
  if [ ! -x dist/pg-helper ]; then
    echo >&2 "pg-helper not built yet"
    exit 1
  fi
  rsync -av dist e2e/

[no-cd,private]
_clean_pg_helper_in_e2e:
  rm -rf e2e/dist
