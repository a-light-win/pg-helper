[no-cd,private]
@secrets-dir:
  mkdir -p tests/secrets

[no-cd,private]
@pg-admin-password: secrets-dir
  echo "Generating pg admin password"
  echo "post\ngres" > tests/secrets/pg_admin_password

[no-cd,private]
@clean-pg-admin-password:
  echo "Cleaning pg admin password"
  rm -rf tests/secrets/pg_admin_password

[no-cd,private]
gen-test-jwt-key: secrets-dir
  #!/usr/bin/env bash
  if [ ! -e "tests/secrets/auth_private_key" ]; then
    echo "Generating jwt auth key for testing"
    ./dist/pg-helper gen-key --config tests/config.yaml
    chmod 644 "tests/secrets/auth_private_key" "tests/secrets/auth_public_key"
  fi

[no-cd,private]
@clean-test-jwt-key:
  echo "Cleaning jwt auth key for testing"
  rm -rf "tests/secrets/auth_private_key" "tests/secrets/auth_public_key"

[no-cd,private]
@gen-test-jwt name scopes resources: secrets-dir
  echo "Generating JWT for {{ name }} with scope {{ scopes }} and resources {{ resources }}"
  ./dist/pg-helper gen-jwt --config tests/config.yaml \
    --output "tests/secrets/auth_token_{{ replace(name, '-', '_') }}" \
    --subject "{{ name }}" \
    --scopes "{{ scopes }}" \
    --resources "{{ resources }}"

  chmod 644 "tests/secrets/auth_token_{{ replace(name, '-', '_') }}"

[no-cd,private]
@clean-test-jwt name:
  echo "Cleaning JWT for {{ name }}"
  rm -rf "tests/secrets/auth_token_{{ replace(name, '-', '_') }}"

[no-cd]
gen-all: \
  pg-admin-password \
  gen-test-jwt-key \
  (gen-test-jwt "web" "db:read,db:write" "db:test1,db:test2") \
  (gen-test-jwt "pg_waiter" "db:read" "db") \
  (gen-test-jwt "grpc_agent" "agent" "dbInstance")


[no-cd]
clean-all: \
  clean-pg-admin-password \
  clean-test-jwt-key \
  (clean-test-jwt "web") \
  (clean-test-jwt "pg_waiter") \
  (clean-test-jwt "grpc_agent")
