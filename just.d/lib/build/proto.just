
[no-cd,private]
_install_protoc:
  #!/usr/bin/env bash
  which protoc &> /dev/null
  if [ $? -eq 0 ]; then
    exit 0
  fi

  echo "Installing protoc ..."
  if which apt-get &> /dev/null ; then
    sudo apt install -y protobuf-compiler
  elif which pacman &> /dev/null ; then
    sudo pacman -S protobuf
  fi 

[no-cd,private]
_install_protoc_gen_go: _install_protoc
  #!/usr/bin/env bash
  which protoc-gen-go &> /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing protoc-gen-go ..."
    GOBIN=~/.local/bin go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  fi

[no-cd,private]
_install_protoc_gen_go_grpc: _install_protoc_gen_go
  #!/usr/bin/env bash
  which protoc-gen-go-grpc &> /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing protoc-gen-go-grpc ..."
    GOBIN=~/.local/bin go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  fi

[no-cd,private]
_generate_protos: _install_protoc_gen_go_grpc
  #!/usr/bin/env bash
  echo "Generating protos ..."
  protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/proto/*.proto

[no-cd,private]
_clean_protos:
  rm -rf pkg/proto/*.pb.go

