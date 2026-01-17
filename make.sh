#!/usr/bin/env sh
set -e

IMG_DEFAULT="ivannik2910/eirenyx:latest"
IMG="${IMG:-$IMG_DEFAULT}"

cmd="$1"

usage() {
  echo "Usage: ./make.sh <command>"
  echo ""
  echo "Commands:"
  echo "  dev          Generate + manifests + run locally"
  echo "  api          Regenerate APIs and CRDs"
  echo "  build        Build controller binary"
  echo "  image        Build docker image"
  echo "  push         Push docker image"
  echo "  deploy       Deploy controller to cluster"
  echo "  undeploy     Remove controller from cluster"
  echo "  install-crd  Install CRDs"
  echo "  uninstall-crd Uninstall CRDs"
  echo "  test         Run unit tests"
  echo ""
  echo "Environment variables:"
  echo "  IMG=<image>  Docker image (default: $IMG_DEFAULT)"
}

case "$cmd" in
  dev)
    make generate manifests run
    ;;
  api)
    make generate manifests
    ;;
  build)
    make build
    ;;
  image)
    make docker-build IMG="$IMG"
    ;;
  push)
    make docker-push IMG="$IMG"
    ;;
  deploy)
    make deploy IMG="$IMG"
    ;;
  undeploy)
    make undeploy
    ;;
  install-crd)
    make install
    ;;
  uninstall-crd)
    make uninstall
    ;;
  test)
    make test
    ;;
  *)
    usage
    exit 1
    ;;
esac
