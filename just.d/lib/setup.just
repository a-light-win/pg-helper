
import './vars.just'

[no-cd,private]
_update_repo_cache:
  #!/usr/bin/env bash
  if which pacman &>/dev/null ; then
    {{ sudo_cmd }} pacman -Sy
  elif which apt &>/dev/null ; then
    {{ sudo_cmd }} apt update
  fi

[no-cd,private]
_setup name ignore_error="false":
  #!/usr/bin/env bash
  ret=1
  if which pacman &>/dev/null ; then
    {{ sudo_cmd }} pacman -S --noconfirm --needed {{ name }}
    ret=$?
  elif which apt &>/dev/null ; then
    {{ sudo_cmd }} apt install -y {{ name }}
    ret=$?
  fi

  if [ $ret -ne 0 ] && [ "{{ ignore_error }}" != "true" ]; then
    echo "Failed to install {{ name }}"
    exit 1
  fi

# Only install the package if the package manager is available
[no-cd,private]
_setup_with pkg_manager name:
  #!/usr/bin/env bash
  if [ "{{ pkg_manager }}" = pacman ]; then
    if which pacman &>/dev/null ; then
      {{ sudo_cmd }} pacman -S --noconfirm --needed {{ name }}
    fi
  elif [ "{{ pkg_manager }}" = apt ]; then
    if which apt &>/dev/null ; then
      {{ sudo_cmd }} apt install -y {{ name }}
    fi
  fi


_setup_protoc: (_setup_with 'pacman' 'protobuf') \
  (_setup_with 'apt' 'protobuf-compiler')

[no-cd,private]
_setup_protoc_gen_go:
  #!/usr/bin/env bash
  which protoc-gen-go &> /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing protoc-gen-go {{ protoc_gen_go_version }}..."
  else
    existing_version=$(protoc-gen-go --version|awk '{print $2}')
    if [ "$existing_version" = "{{ protoc_gen_go_version }}" ]; then
      echo "protoc-gen-go {{protoc_gen_go_version}} is already installed"
      exit 0
    fi
    echo "Upgrading protoc-gen-go to {{ protoc_gen_go_version }}..."
  fi

  GOBIN={{ gobin_path }} go install google.golang.org/protobuf/cmd/protoc-gen-go@{{ protoc_gen_go_version }}

[no-cd,private]
_setup_protoc_gen_go_grpc:
  #!/usr/bin/env bash
  which protoc-gen-go-grpc &> /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing protoc-gen-go-grpc {{ protoc_gen_go_grpc_version }} ..."
  else
    existing_version=$(protoc-gen-go-grpc --version|awk '{print $2}')
    if [ v"$existing_version" = "{{ protoc_gen_go_grpc_version }}" ]; then
      echo "protoc-gen-go-grpc {{protoc_gen_go_grpc_version}} is already installed"
      exit 0
    fi
    echo "Upgrading protoc-gen-go-grpc to {{ protoc_gen_go_grpc_version }}..."
  fi

  GOBIN={{ gobin_path }} go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@{{ protoc_gen_go_grpc_version }}

[no-cd,private]
_setup_sqlc:
  #!/usr/bin/env bash
  which sqlc &> /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing sqlc ..."
  else
    existing_version=$(sqlc version)
    if [ "$existing_version" = "{{ sqlc_version }}" ]; then
      echo "sqlc {{sqlc_version}} is already installed"
      exit 0
    fi
    echo "Upgrading sqlc to {{ sqlc_version }}..."
  fi

  GOBIN={{ gobin_path }} go install github.com/sqlc-dev/sqlc/cmd/sqlc@{{ sqlc_version }}

[no-cd,private]
_setup_goose:
  #!/usr/bin/env bash
  which goose &> /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing goose ..."
  else
  existing_version=$(goose --version|awk '{print $3}')
    if [ "$existing_version" = "{{ goose_version }}" ]; then
      echo "goose {{goose_version}} is already installed"
      exit 0
    fi
    echo "Upgrading goose to {{ goose_version }}..."
  fi
  GOBIN={{ gobin_path }} go install github.com/pressly/goose/v3/cmd/goose@{{ goose_version }}

[no-cd,private]
_setup_venom:
  #!/usr/bin/env bash
  which venom &> /dev/null
  if [ $? -eq 0 ]; then
    existing_version=$(venom version|awk '{print $3}')
    if [ "$existing_version" = "{{ venom_version }}" ]; then
      echo "venom {{venom_version}} is already installed"
      exit 0
    fi
    echo "Upgrading venom to {{ venom_version }}..."
  else
    echo "Installing venom {{ venom_version }}..."
  fi

  arch=$(uname -m)
  if [ "$arch" = "aarch64" ]; then
    arch=arm64
  elif [ "$arch" = "x86_64" ]; then
    arch=amd64
  fi
  curl -L -o "{{ gobin_path / 'venom' }}" https://github.com/ovh/venom/releases/download/{{ venom_version }}/venom.linux-$arch || exit $?
  chmod +x "{{ gobin_path / 'venom' }}"

[no-cd,private]
_setup_podman: (_setup 'buildah podman')

[no-cd,private]
_setup_podman_compose: (_setup 'podman-compose' "true")
  #!/usr/bin/env bash
  grep -q "VERSION_CODENAME=jammy" /etc/os-release
  if [ $? -eq 0 ]; then
    curl -o /usr/local/bin/podman-compose https://raw.githubusercontent.com/containers/podman-compose/main/podman_compose.py
    chmod +x /usr/local/bin/podman-compose
  fi

[no-cd,private]
_setup_proto: (_setup_protoc) (_setup_protoc_gen_go) (_setup_protoc_gen_go_grpc)
