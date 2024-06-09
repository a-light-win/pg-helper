
[no-cd,private]
_generate_protos:
  #!/usr/bin/env bash
  echo "Generating protos ..."
  protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/proto/*.proto

[no-cd,private]
_clean_protos:
  rm -rf pkg/proto/*.pb.go
