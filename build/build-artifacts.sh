#!/bin/bash
set -o errexit

realpath() {
  echo "$(cd "$(dirname "$1")"; pwd)/$(basename "$1")"
}

BUILD_DIR=$(dirname $(realpath $0))
PROJECT_DIR=$(dirname $BUILD_DIR)

setup_output_dir() {
  echo "Creating output directory $1"
  mkdir -p $1
}


run_build() {
  XTRALAB_VERSION=$1
  OUTPUT_DIR=$(realpath $2)

  setup_output_dir ${OUTPUT_DIR}

  docker run \
      --rm \
      -e SCRATCH_DIR="/scratch" \
      -e OUTPUT_DIR="/dist" \
      -e XTRALAB_VERSION="${XTRALAB_VERSION}" \
      -v
}

case $# in
  2)
    run_build $1 $2
    ;;

  *)
    echo "Usage: $0 <version_string> <output-directory> "
    echo "  "
    echo "Example:"
    echo "  ./build-artifacts.sh 0.1.0 ."
    echo "  "
    exit 1
    ;;
esac