#!/usr/bin/env sh
set -e

IMG_DEFAULT="ivannik2910/eirenyx:latest"
IMG="${IMG:-$IMG_DEFAULT}"
KIND_NODE="${KIND_NODE:-eirenyx-dev-control-plane}"

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
  echo "  prune-images Remove unused images from the kind node to reclaim disk"
  echo ""
  echo "Environment variables:"
  echo "  IMG=<image>        Docker image (default: $IMG_DEFAULT)"
  echo "  KIND_NODE=<name>   kind node container name (default: $KIND_NODE)"
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
    make docker-buildx IMG="$IMG"
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
  prune-images)
    echo "Pruning unused images on kind node '$KIND_NODE'..."
    docker exec "$KIND_NODE" crictl rmi --prune
    echo ""
    docker exec "$KIND_NODE" df -h /
    ;;
  *)
    usage
    exit 1
    ;;
esac
