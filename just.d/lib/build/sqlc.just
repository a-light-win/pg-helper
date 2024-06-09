
[no-cd,private]
_sqlc:
  sqlc generate 

[no-cd,private]
_clean_sqlc:
  rm -rf internal/db/db.go \
    internal/db/models.go \
    internal/db/*.sql.go
