import './e2e/secrets.just'
import './e2e/pg-helper.just'
import './e2e/venom.just'

serve_compose_args := "--profile pg-13 --profile pg-14 --profile authelia -f " + container_provider + "-compose.yaml"

[no-cd,private]
_e2e_clean_all: _clean_pg_helper_in_e2e _clean_venom _clean_all_secrets

[no-cd,private]
_serve: _all_secrets _copy_pg_helper_to_e2e
  #!/usr/bin/env bash
  cd e2e
  export PG_HELPER_VERSION=$(./dist/pg-helper version|awk '{print $2}')
  provider={{ if container_provider == 'nerdctl' { 'nerdctl compose' } else { container_provider + '-compose' } }}
  ${provider} {{ serve_compose_args }} up
  ${provider} {{ serve_compose_args }} stop

[no-cd,private]
_clean_serve:
  #!/usr/bin/env bash
  cd e2e
  export PG_HELPER_VERSION=$(./dist/pg-helper version|awk '{print $2}')
  provider={{ if container_provider == 'nerdctl' { 'nerdctl compose' } else { container_provider + '-compose' } }}
  ${provider} {{ serve_compose_args }} down
