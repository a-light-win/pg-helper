
get_cmd := 'curl -Lv -X GET -H "Authorization: Bearer $(< tests/secrets/auth_token_web)" http://127.0.0.1:8080/api/v1/db'
post_cmd := 'curl -L -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $(< tests/secrets/auth_token_web)" http://127.0.0.1:8080/api/v1/db'

[no-cd]
ready db_name:
	{{ get_cmd }}/ready?'db_name={{ db_name }}'

[no-cd]
create db_name:
	{{ post_cmd }} -d '{"dbname": "{{ db_name }}"}'


