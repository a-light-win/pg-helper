#!/bin/sh
trap 'exit 0' TERM INT

token=$(cat "$PG_HELPER_TOKEN")
check_db_ready() {
	if [ -e /db-ready ]; then
		return
	fi

	db_status=$(curl -L --no-progress-meter -X GET \
		-H "Authorization: Bearer ${token}" \
		"${PG_HELPER_URL}/api/v1/db/ready?db_name=${DB_NAME}&&name=${PG_INSTANCE}")

	echo "$db_status" | grep "true" >/dev/null
	if [ $? -eq 0 ]; then
		echo "$(date): The database ${DB_NAME} on ${PG_INSTANCE} is ready."
		touch /db-ready
	fi
}

rm -rf /db-ready

while :; do
	check_db_ready
	if [ -e /db-ready ]; then
		break
	fi
	sleep 3
done

while :; do
	sleep 60
done
