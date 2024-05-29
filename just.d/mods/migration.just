
[no-cd,private]
install-goose:
  #!/usr/bin/env bash
  which goose &> /dev/null
  if [ $? -ne 0 ]; then
    echo "Installing goose ..."
    go install github.com/pressly/goose/v3/cmd/goose@latest
  fi

[no-cd]
new name: install-goose
  goose -dir internal/db/migrations create {{ name }} sql
