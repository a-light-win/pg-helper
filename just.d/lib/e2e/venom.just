
[no-cd,private]
_venom compose_provider name: _copy_pg_helper_to_e2e
  #!/usr/bin/env bash
  
  cd e2e

  export PG_HELPER_VERSION=$(./dist/pg-helper version | awk '{print $2}')

  output_dir="output/{{compose_provider}}-{{ name }}"
  venom run --output-dir "$output_dir" --var-from-file "tests/var/compose-provider-{{ compose_provider }}.yaml" --var-from-file "tests/var/{{name}}.yaml" --lib-dir=tests/lib "tests/{{name}}.yaml"

# clean the output of venom
[no-cd,private]
_clean_venom:
  rm -rf e2e/output
