#!/usr/bin/env bash

set -e

ROOT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/.."
BUILD_DIR="$ROOT_DIR/build"

make clean
make all-cross

mv -f "$BUILD_DIR/drelayer-linux-x64" "$BUILD_DIR/drelayer"
tar -czvf "$BUILD_DIR/drelayer-linux-x64.tgz" -C "$BUILD_DIR" "drelayer"

mv -f "$BUILD_DIR/drelayer-darwin-x64" "$BUILD_DIR/drelayer"
tar -czvf "$BUILD_DIR/drelayer-darwin-x64.tgz" -C "$BUILD_DIR" "drelayer"

gpg2 --detach-sig --default-key D4B604F1 --output "$BUILD_DIR/drelayer-linux-x64.tgz.sig" "$BUILD_DIR/drelayer-linux-x64.tgz"
gpg2 --detach-sig --default-key D4B604F1 --output "$BUILD_DIR/drelayer-darwin-x64.tgz.sig" "$BUILD_DIR/drelayer-darwin-x64.tgz"

s3cmd put "$BUILD_DIR/drelayer-linux-x64.tgz" s3://ddrp-releases/drelayer-linux-x64.tgz
s3cmd setacl s3://ddrp-releases/drelayer-linux-x64.tgz --acl-public
s3cmd put "$BUILD_DIR/drelayer-linux-x64.tgz.sig" s3://ddrp-releases/drelayer-linux-x64.tgz.sig
s3cmd setacl s3://ddrp-releases/drelayer-linux-x64.tgz.sig --acl-public
cd "$BUILD_DIR" && shasum -a 256 drelayer-linux-x64.tgz > /tmp/drelayer-linux-x64.tgz.sum.txt && cd "$DIR"
s3cmd put /tmp/drelayer-linux-x64.tgz.sum.txt s3://ddrp-releases/drelayer-linux-x64.tgz.sum.txt
s3cmd setacl s3://ddrp-releases/drelayer-linux-x64.tgz.sum.txt --acl-public