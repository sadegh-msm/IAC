#!/bin/bash

set -e

IMG_BASE="sadegh81/monogdb-operator"

function usage() {
  echo "Usage:"
  echo "  $0 install <tag>    # Build, push, install, apply samples, deploy"
  echo "  $0 delete           # Delete samples, uninstall, undeploy"
  exit 1
}

function install() {
  local tag="$1"
  if [ -z "$tag" ]; then
    echo "Error: Tag is required for install."
    usage
  fi

  cd certs

  kubectl create ns mongo-operator-system
  kubectl create secret tls webhook-server-cert --cert=tls.crt --key=tls.key -n mongo-operator-system

  cd ..

  local img="${IMG_BASE}:${tag}"

  echo ">>> Building and pushing Docker image: $img"
  make docker-build docker-push IMG="$img"

  echo ">>> Installing CRDs"
  make install

  echo ">>> Applying sample resources"
  kubectl apply -k config/samples/

  echo ">>> Deploying operator"
  make deploy IMG="$img"

  echo "✅ Install complete."
}

function delete_resources() {
  echo ">>> Deleting sample resources"
  kubectl delete -k config/samples/ || true

  echo ">>> Uninstalling CRDs"
  make uninstall || true

  echo ">>> Undeploying operator"
  make undeploy || true

  echo "✅ Delete complete."
}

# Main dispatcher
case "$1" in
  install)
    install "$2"
    ;;
  delete)
    delete_resources
    ;;
  *)
    usage
    ;;
esac

