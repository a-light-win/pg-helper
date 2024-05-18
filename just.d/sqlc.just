
[no-cd,private]
sqlc:
  {{ env('DOCKER_CMD', 'podman') }} run -it --rm -v `pwd`:`pwd` -w `pwd` docker.io/sqlc/sqlc generate 

[no-cd,private]
clean-sqlc:
  rm -rf internal/db/db.go \
    internal/db/models.go \
    internal/db/*.sql.go

