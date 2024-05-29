import "../build.just"

[no-cd,private]
_copy_pg_helper_to_e2e: _build_all
  #!/usr/bin/env bash
  rsync -av dist e2e/

[no-cd,private]
_clean_pg_helper_in_e2e:
  rm -rf e2e/dist
