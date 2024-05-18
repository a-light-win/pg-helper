
[no-cd,private]
strip name:
  #!/usr/bin/env bash
  echo "Stripping {{ name }} ..."

  cd dist/
  file "{{ name }}" |grep -q "not stripped"
  if [ $? -ne 0 ]; then
    echo "{{ name }} binary already stripped"
    exit 0
  fi

  objcopy --only-keep-debug "{{ name }}" "{{ name }}.dbg"
  objcopy --strip-unneeded "{{ name }}"
  objcopy --add-gnu-debuglink="{{ name }}.dbg" "{{ name }}"

  echo "Stripping {{ name }} success"

