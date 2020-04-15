#!/usr/bin/env bash

set -e

ROOT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/.."
BUILD_DIR="$ROOT_DIR/build"
USAGE="release.sh <git-tag> <release-message>"

if [ -z "$1" ]; then
  echo "Must supply a Git tag name."
  echo "$USAGE"
  exit 1
fi

if [ -z "$2" ]; then
    echo "Must supply a release message."
    echo "$USAGE"
    exit 1
fi

TAG=$1
MESSAGE=$2

make clean
make all-cross

mv -f "$BUILD_DIR/drelayer-linux-x64" "$BUILD_DIR/drelayer"
tar -czvf "$BUILD_DIR/drelayer-linux-x64.tgz" -C "$BUILD_DIR" "drelayer"

mv -f "$BUILD_DIR/drelayer-darwin-x64" "$BUILD_DIR/drelayer"
tar -czvf "$BUILD_DIR/drelayer-darwin-x64.tgz" -C "$BUILD_DIR" "drelayer"

gpg2 --detach-sig --default-key D4B604F1 --output "$BUILD_DIR/drelayer-linux-x64.tgz.sig" "$BUILD_DIR/drelayer-linux-x64.tgz"
gpg2 --detach-sig --default-key D4B604F1 --output "$BUILD_DIR/drelayer-darwin-x64.tgz.sig" "$BUILD_DIR/drelayer-darwin-x64.tgz"

git tag "$TAG" || echo "Tag already exists."
git push origin --tags --force

gothub release --user kyokan --repo ddrp-drelayer --tag "$TAG" --name "$TAG" --description "$MESSAGE"
gothub upload --user kyokan --repo ddrp-drelayer --tag "$TAG" --file "$BUILD_DIR/drelayer-linux-x64.tgz" --name "drelayer-linux-x64.tgz"
gothub upload --user kyokan --repo ddrp-drelayer --tag "$TAG" --file "$BUILD_DIR/drelayer-linux-x64.tgz.sig" --name "drelayer-linux-x64.tgz.sig"
gothub upload --user kyokan --repo ddrp-drelayer --tag "$TAG" --file "$BUILD_DIR/drelayer-darwin-x64.tgz" --name "drelayer-darwin-x64.tgz"
gothub upload --user kyokan --repo ddrp-drelayer --tag "$TAG" --file "$BUILD_DIR/drelayer-darwin-x64.tgz.sig" --name "drelayer-darwin-x64.tgz.sig"

# for seed node deployments

s3cmd put "$BUILD_DIR/drelayer-linux-x64.tgz" s3://ddrp-releases/drelayer-linux-x64.tgz
s3cmd setacl s3://ddrp-releases/drelayer-linux-x64.tgz --acl-public
s3cmd put "$BUILD_DIR/drelayer-linux-x64.tgz.sig" s3://ddrp-releases/drelayer-linux-x64.tgz.sig
s3cmd setacl s3://ddrp-releases/drelayer-linux-x64.tgz.sig --acl-public
cd "$BUILD_DIR" && shasum -a 256 drelayer-linux-x64.tgz > /tmp/drelayer-linux-x64.tgz.sum.txt && cd "$DIR"
s3cmd put /tmp/drelayer-linux-x64.tgz.sum.txt s3://ddrp-releases/drelayer-linux-x64.tgz.sum.txt
s3cmd setacl s3://ddrp-releases/drelayer-linux-x64.tgz.sum.txt --acl-public