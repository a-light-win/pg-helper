
[no-cd,private]
download-venom:
  #!/usr/bin/env bash
  if [ ! -x ~/.local/bin/venom ]; then
    echo "Installing venom..."

    curl https://github.com/ovh/venom/releases/download/v1.2.0/venom.linux-amd64 -L -o ~/.local/bin/venom &>/dev/null || exit 1
    chmod +x ~/.local/bin/venom

    echo "venom installed in ~/.local/bin/venom"
  fi

[no-cd,private]
update-venom: download-venom
  #!/usr/bin/env bash
  echo "Updating venom..."
  out=$(venom update 2>&1)
  if [ $? -eq 0 ] ; then
    echo "venom updated"
    exit 0
  fi

  if echo "$out" | grep -q "already have" ; then
    echo "venom is already up to date"
    exit 0
  fi 

  echo >&2 "Update venom failed: $out"
  exit 1


[no-cd]
test: update-venom

