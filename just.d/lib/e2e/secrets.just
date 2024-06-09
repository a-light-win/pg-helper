
import '../vars.just'

special_chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

[no-cd,private]
_all_secrets: \
  (_password 'pg_helper_e2e_pg_admin_password' special_chars) \
  (_jwt_key 'pg_helper_e2e_pg_helper_jwt_key' 'pg_helper_e2e_pg_helper_jwt_public_key') \
  (_password "pg_helper_e2e_authelia_db_owner") \
  (_password "pg_helper_e2e_test2_db_owner") \
  (_jwt "pg_helper_e2e_grpc_agent" "grpc_agent" "agent" "dbInstance") \
  (_jwt "pg_helper_e2e_authelia_pg_waiter" "pg_waiter" "db:read" "db:authelia") \
  (_password "pg_helper_e2e_authelia_jwt_secret") \
  (_password "pg_helper_e2e_authelia_session_secret") \
  (_password "pg_helper_e2e_authelia_storage_encryption_key")


[no-cd,private]
_clean_all_secrets: \
  (_clean_secret 'pg_helper_e2e_pg_admin_password') \
  (_clean_jwt_key 'pg_helper_e2e_pg_helper_jwt_key' 'pg_helper_e2e_pg_helper_jwt_public_key') \
  (_clean_secret "pg_helper_e2e_authelia_db_owner") \
  (_clean_secret "pg_helper_e2e_test2_db_owner") \
  (_clean_secret "pg_helper_e2e_grpc_agent") \
  (_clean_secret "pg_helper_e2e_authelia_pg_waiter") \
  (_clean_secret "pg_helper_e2e_authelia_jwt_secret") \
  (_clean_secret "pg_helper_e2e_authelia_session_secret") \
  (_clean_secret "pg_helper_e2e_authelia_storage_encryption_key")

[no-cd,private]
@_secrets_dir:
  mkdir -p {{ secrets_dir }}

[no-cd,private]
_password secret_name charset=HEX: _secrets_dir && (_compose_secret secret_name)
  #!/usr/bin/env bash
  echo "Generating password for {{ secret_name }}"

  _password_file="{{ secrets_dir / secret_name }}"
  if [ -e "${_password_file}" ]; then
    echo "Password for {{ secret_name }} already exists"
    exit 0
  fi
  echo "{{ choose('32', HEX) }}" > "${_password_file}"
  chmod 640 "${_password_file}"

[no-cd,private]
_jwt_key private_key_name public_key_name: _secrets_dir && (_compose_secret public_key_name)
  #!/usr/bin/env bash
  private_key_file="{{ secrets_dir / private_key_name }}"
  public_key_file="{{ secrets_dir / public_key_name }}"
  if [ ! -e "${private_key_file}" ]; then
    echo "Generating jwt auth key for {{ private_key_name }}"
    ./dist/pg-helper gen-key --config e2e/config/local/pg-helper.yaml
  fi

[no-cd,private]
_jwt secret_name name scopes resources: _secrets_dir && (_compose_secret secret_name)
  #!/usr/bin/env bash
  echo "Generating JWT {{ secret_name }} with scope {{ scopes }} and resources {{ resources }}"
  _jwt_file="{{ secrets_dir / secret_name }}"
  ./dist/pg-helper gen-jwt --config e2e/config/local/pg-helper.yaml \
    --output "${_jwt_file}" \
    --subject "{{ name }}" \
    --scopes "{{ scopes }}" \
    --resources "{{ resources }}"

[no-cd,private]
_ensure_compose_provider_exists:
  #!/usr/bin/env bash
  echo "docker podman nerdctl" | grep -w "{{ container_provider }}" > /dev/null || (echo >&2 "Invalid compose provider" && exit 1)

[no-cd,private]
_compose_secret secret_name: _ensure_compose_provider_exists
  #!/usr/bin/env bash
  exists_name=$({{container_provider }} secret ls --format '{{{{.Name}}'|grep -w "{{ secret_name }}")
  if [ -n "$exists_name" ]; then
    echo "Secret {{ secret_name }} already exists"
    exit 0
  fi
  {{ container_provider }} secret create "{{ secret_name }}" "{{ secrets_dir / secret_name }}"

[no-cd,private]
_clean_jwt_key private_key_name public_key_name: (_clean_secret public_key_name)
  #!/usr/bin/env bash
  echo "Cleaning secret for {{ private_key_name }}"
  private_key="{{ secrets_dir / private_key_name }}"
  rm -rf "${private_key}"

[no-cd,private]
_clean_secret secret_name: (_clean_compose_secret secret_name)
  #!/usr/bin/env bash
  echo "Cleaning secret for {{ secret_name }}"
  secret_file="{{ secrets_dir / secret_name }}"
  rm -rf "${secret_file}"

[no-cd,private]
_clean_compose_secret secret_name: _ensure_compose_provider_exists
  #!/usr/bin/env bash
  if {{ container_provider }} secret exists "{{ secret_name }}"; then
    {{ container_provider }} secret rm "{{ secret_name }}"
  fi
