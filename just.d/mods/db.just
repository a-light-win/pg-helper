
get_cmd := 'curl -Lv -X GET -H "Authorization: Bearer $(< tests/secrets/auth_token_web)" http://127.0.0.1:8080/api/v1/db'
post_cmd := 'curl -L -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $(< tests/secrets/auth_token_web)" http://127.0.0.1:8080/api/v1/db'

[no-cd]
ready db_name instance='pg-13':
	{{ get_cmd }}/ready?'db_name={{ db_name }}&&name={{ instance }}'

[no-cd]
create db_name instance='pg-13': (create-db-parameters db_name instance '')
	{{ post_cmd }} -d "@tests/secrets/create-db-{{ db_name }}"

[no-cd]
migrate db_name instance='pg-14' from='pg-13': (create-db-parameters db_name instance from)
	{{ post_cmd }} -d "@tests/secrets/create-db-{{ db_name }}"

[no-cd,private]
create-db-parameters db_name instance from:
	#!/usr/bin/env bash
	cat > "tests/secrets/create-db-{{ db_name }}" <<EOF
	{
		"name": "{{ instance }}",
		"db_name": "{{ db_name }}",
		"db_owner": "{{ db_name }}",
		"db_password": "{{ db_name}}-test",
	EOF

	if [[ -n "{{ from }}" ]]; then
		echo '        "migrate_from": "{{ from }}",' >> "tests/secrets/create-db-{{ db_name }}"
	fi 

	cat >> "tests/secrets/create-db-{{ db_name }}" <<EOF
		"reason": "test"
	}
	EOF
