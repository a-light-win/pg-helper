
[no-cd,private]
install-protoc:
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
install-protoc-gen-go: install-protoc
  #!/usr/bin/env bash
  which protoc-gen-go &> /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing protoc-gen-go ..."
    GOBIN=~/.local/bin go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  fi

[no-cd,private]
install-protoc-gen-go-grpc: install-protoc-gen-go
  #!/usr/bin/env bash
  which protoc-gen-go-grpc &> /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing protoc-gen-go-grpc ..."
    GOBIN=~/.local/bin go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  fi

[no-cd,private]
generate-protos: install-protoc-gen-go-grpc
  #!/usr/bin/env bash
  echo "Generating protos ..."
  protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/proto/*.proto

[no-cd,private]
clean-protos:
  rm -rf api/proto/*.pb.go

