#!/bin/sh

token=$(cat "$PG_HELPER_TOKEN_FILE")
db_name="$1"
instance_name="$2"

echo "$(date): Waiting for the database ${db_name} on ${instance_name} to be ready..."

check_db_ready() {
	db_name="$1"
	instance_name="$2"
	db_status=$(curl -L --no-progress-meter -X GET \
		-H "Authorization: Bearer ${token}" \
		"${PG_HELPER_URL}/api/v1/db/ready?name=${db_name}&&instance_name=${instance_name}")

	echo "$db_status" | grep "true" >/dev/null
	if [ $? -eq 0 ]; then
		echo "$(date): The database ${db_name} on ${instance_name} is ready."
		return 0
	fi
	return 1
}

while :; do
	check_db_ready "$db_name" "$instance_name"
	if [ $? -eq 0 ]; then
		break
	fi
	sleep 3
done

shift 2
"$@"
