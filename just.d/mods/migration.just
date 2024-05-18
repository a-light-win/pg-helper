
[no-cd,private]
install-goose:
  #!/usr/bin/env bash
  which goose &> /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing goose ..."
    GOBIN=~/.local/bin go install github.com/pressly/goose/v3/cmd/goose@latest
  fi

new name: install-goose
  goose -dir db/migrations create {{ name }} sql
