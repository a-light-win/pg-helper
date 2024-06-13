#!/usr/bin/env bash

just_version="$1"
arch=$(uname -m)
if [ "$arch" = "amd64" ]; then
	arch="x86_64"
elif [ "$arch" = "arm64" ]; then
	arch="aarch64"
fi

just_pkg="just-${just_version}-${arch}-unknown-linux-musl.tar.gz"

cd /tmp/
curl -o "$just_pkg" -L "https://github.com/casey/just/releases/download/${just_version}/${just_pkg}"

cd /usr/local/bin
tar --transform 's/LICENSE/just-license/' -xvf "/tmp/${just_pkg}" just LICENSE

rm -rf "/tmp/${just_pkg}"
