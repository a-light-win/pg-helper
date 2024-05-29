
mod secrets "just.d/mods/secrets.just"
mod migration "just.d/mods/migration.just"
mod db "just.d/mods/db.just"
mod e2e "just.d/mods/e2e.just"
mod build "just.d/mods/build.just"
mod test "just.d/mods/test.just"

serve_compose_args := "--profile pg-13 --profile pg-14 --profile authelia -f " + env('DOCKER_CMD', 'podman') + "-compose.yaml"

#serve: build
#  #!/usr/bin/env bash
#  export PG_HELPER_VERSION=$(./dist/pg-helper version|awk '{print $2}')
#  {{ env('DOCKER_CMD', 'podman')}}-compose {{ serve_compose_args }} up --force-recreate --build
#  {{ env('DOCKER_CMD', 'podman')}}-compose {{ serve_compose_args }} stop
#
#clean-serve:
#  #!/usr/bin/env bash
#  set -x
#  export PG_HELPER_VERSION=$(./dist/pg-helper version|awk '{print $2}')
#  {{ env('DOCKER_CMD', 'podman')}}-compose {{ serve_compose_args }} down
#  
#clean: clean-sqlc clean-protos clean-serve
#  rm -rf dist/
