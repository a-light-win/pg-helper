
[no-cd,private]
_download_venom:
  #!/usr/bin/env bash
  if [ ! -x ~/go/bin/venom ]; then
    echo "Installing venom..."

    go install -v github.com/ovh/venom@latest || exit 1

    echo "venom installed in ~/go/bin/venom"
  fi

[no-cd,private]
_update_venom: _download_venom
  #!/usr/bin/env bash
  echo "Updating venom..."
  out=$(venom update 2>&1)
  if [ $? -eq 0 ] ; then
    echo "venom updated"
    exit 0
  fi

  if echo "$out" | grep -q "already have" ; then
    echo "venom is already up to date"
    exit 0
  fi 

  echo >&2 "Update venom failed: $out"
  exit 1

[no-cd,private]
_venom name: _copy_pg_helper_to_e2e
  #!/usr/bin/env bash
  
  cd e2e

  export PG_HELPER_VERSION=$(./dist/pg-helper version | awk '{print $2}')

  output_dir=output/{{ if name == '*' { 'all' } else { name } }}

  venom run --output-dir "$output_dir" --lib-dir=tests/lib tests/{{ name }}.yaml

# clean the output of venom
[no-cd,private]
_clean_venom:
  rm -rf e2e/output
