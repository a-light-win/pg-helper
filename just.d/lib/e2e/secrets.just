secrets_dir := 'e2e/secrets'


[no-cd,private]
_all_secrets: \
  _pg_admin_password \
  _jwt_key \
  (_password "pg-helper" "db_owner_authelia") \
  (_password "pg-helper" "db_owner_test2") \
  (_jwt "pg-helper-agent" "grpc_agent" "agent" "dbInstance") \
  (_jwt "local" "web" "db:read,db:write" "db:test1,db:test2") \
  (_jwt "authelia" "pg_waiter" "db:read" "db:authelia") \
  (_password "authelia" "jwt_secret") \
  (_password "authelia" "session_secret") \
  (_password "authelia" "storage_encryption_key")


[no-cd,private]
_clean_all_secrets: \
  _clean_pg_admin_password \
  _clean_jwt_key \
  (_clean_password "pg-helper" "db_owner_authelia") \
  (_clean_password "pg-helper" "db_owner_test2") \
  (_clean_jwt "pg-helper-agent" "grpc_agent") \
  (_clean_jwt "local" "web") \
  (_clean_jwt "authelia" "pg_waiter") \
  (_clean_password "authelia" "jwt_secret") \
  (_clean_password "authelia" "session_secret") \
  (_clean_password "authelia" "storage_encryption_key")

[no-cd,private]
@_secrets_dir app_name:
  mkdir -p {{ secrets_dir / app_name }}

[no-cd,private]
@_password app_name secret_name: (_secrets_dir app_name)
  echo "Generating password for {{ app_name / secret_name }}"
  echo "{{ choose('32', HEX) }}" > {{ secrets_dir / app_name / secret_name }}
  chmod 640 {{ secrets_dir / app_name / secret_name }}

[no-cd,private]
@_clean_password app_name secret_name:
  echo "Cleaning password for {{ app_name / secret_name }}"
  rm -rf {{ secrets_dir / app_name / secret_name }}

[no-cd,private]
@_pg_admin_password: (_secrets_dir 'common')
  echo "Generating pg admin password"
  echo "post\ngres" > {{ secrets_dir / 'common/pg_admin_password' }}
  chmod 640 {{ secrets_dir / 'common/pg_admin_password' }}

[no-cd,private]
@_clean_pg_admin_password: (_clean_password 'common' 'pg_admin_password')

[no-cd,private]
_jwt_key: (_secrets_dir 'local') (_secrets_dir 'pg-helper')
  #!/usr/bin/env bash
  private_key="{{ secrets_dir / 'local/auth_private_key' }}"
  if [ ! -e "${private_key}" ]; then
    echo "Generating jwt auth key for testing"
    ./dist/pg-helper gen-key --config e2e/config/local/pg-helper.yaml
  fi

[no-cd,private]
_clean_jwt_key:
  #!/usr/bin/env bash
  echo "Cleaning jwt auth key for testing"
  private_key="{{ secrets_dir / 'local/auth_private_key' }}"
  public_key="{{ secrets_dir / 'pg-helper/auth_public_key' }}"
  rm -rf "${private_key}" "${public_key}"

[no-cd,private]
_jwt app_name name scopes resources: (_secrets_dir app_name)
  #!/usr/bin/env bash
  echo "Generating JWT for {{app_name}}/{{ name }} with scope {{ scopes }} and resources {{ resources }}"
  _jwt_file="{{ secrets_dir / app_name / 'auth_token_' + replace(name, '-', '_') }}"
  ./dist/pg-helper gen-jwt --config e2e/config/local/pg-helper.yaml \
    --output "${_jwt_file}" \
    --subject "{{ name }}" \
    --scopes "{{ scopes }}" \
    --resources "{{ resources }}"

[no-cd,private]
_clean_jwt app_name name:
  #!/usr/bin/env bash
  echo "Cleaning JWT for {{app_name}}/{{ name }}"
  _jwt_file="{{ secrets_dir / app_name / 'auth_token_' + replace(name, '-', '_') }}"
  rm -rf "${_jwt_file}"
