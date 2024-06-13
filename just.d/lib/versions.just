
protoc_gen_go_version := `sed -n 's%^\s*google.golang.org/protobuf\s\+\([v0-9.]\+\)$%\1%p' go.mod`
protoc_gen_go_grpc_version := "v1.4.0"
sqlc_version := `sed -n 's%^\s*github.com/sqlc-dev/sqlc\s\+\([v0-9.]\+\)$%\1%p' just.d/lib/go.mod`
goose_version := `sed -n 's%^\s*github.com/pressly/goose/v[0-9]*\s\+\([v0-9.]\+\)$%\1%p' go.mod`
venom_version := `sed -n 's%^\s*github.com/ovh/venom\s\+\([v0-9.]\+\)$%\1%p' just.d/lib/go.mod`
