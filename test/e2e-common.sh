#!/bin/bash

# Copyright 2020 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script includes common functions for testing setup and teardown.

export GO111MODULE=on

source $(dirname $0)/../vendor/knative.dev/test-infra/scripts/e2e-tests.sh

# If gcloud is not available make it a no-op, not an error.
which gcloud &>/dev/null || gcloud() { echo "[ignore-gcloud $*]" 1>&2; }

# Use GNU tools on macOS. Requires the 'grep' and 'gnu-sed' Homebrew formulae.
if [ "$(uname)" == "Darwin" ]; then
  sed=gsed
  grep=ggrep
fi

# Discovery main config.
readonly DISCOVERY_CONFIG="config/"

# The number of controlplane replicas to run.
readonly REPLICAS=3

TMP_DIR=$(mktemp -d -t ci-$(date +%Y-%m-%d-%H-%M-%S)-XXXXXXXXXX)
readonly TMP_DIR
readonly KNATIVE_DEFAULT_NAMESPACE="knative-discovery"

# This the namespace used to install and test Knative Discovery.
export TEST_DISCOVERY_NAMESPACE
TEST_DISCOVERY_NAMESPACE="${TEST_DISCOVERY_NAMESPACE:-"knative-discovery-"$(cat /dev/urandom \
  | tr -dc 'a-z0-9' | fold -w 10 | head -n 1)}"

latest_version() {
  local semver=$(git describe --match "v[0-9]*" --abbrev=0)
  local major_minor=$(echo "$semver" | cut -d. -f1-2)

  # Get the latest patch release for the major minor
  git tag -l "${major_minor}*" | sort -r --version-sort | head -n1
}

# Latest release. If user does not supply this as a flag, the latest
# tagged release on the current branch will be used.
readonly LATEST_RELEASE_VERSION=$(latest_version)

UNINSTALL_LIST=()

# Setup the Knative environment for running tests.
function knative_setup() {
  install_knative_discovery
  unleash_quacken || fail_test "Could not unleash the quacken"
}

function scale_controlplane() {
  for deployment in "$@"; do
    # Make sure all pods run in leader-elected mode.
    kubectl -n "${TEST_DISCOVERY_NAMESPACE}" scale deployment "$deployment" --replicas=0 || failed=1
    # Give it time to kill the pods.
    sleep 5
    # Scale up components for HA tests
    kubectl -n "${TEST_DISCOVERY_NAMESPACE}" scale deployment "$deployment" --replicas="${REPLICAS}" || failed=1
  done
}

# This installs everything from the config dir but then removes the Channel Based Broker.
# TODO: This should only install the core.
# Args:
#  - $1 - if passed, it will be used as discovery config directory
function install_knative_discovery() {
  echo ">> Creating ${TEST_DISCOVERY_NAMESPACE} namespace if it does not exist"
  kubectl get ns ${TEST_DISCOVERY_NAMESPACE} || kubectl create namespace ${TEST_DISCOVERY_NAMESPACE}
  local knd_config
  knd_config="${1:-${DISCOVERY_CONFIG}}"
  # Install Knative Discovery in the current cluster.
  echo "Installing Knative Discovery from: ${knd_config}"
  if [ -d "${knd_config}" ]; then
    local TMP_CONFIG_DIR=${TMP_DIR}/config
    mkdir -p ${TMP_CONFIG_DIR}
    cp -r ${knd_config}/* ${TMP_CONFIG_DIR}
    find ${TMP_CONFIG_DIR} -type f -name "*.yaml" -exec sed -i "s/namespace: ${KNATIVE_DEFAULT_NAMESPACE}/namespace: ${TEST_DISCOVERY_NAMESPACE}/g" {} +
    ko apply --strict -f "${TMP_CONFIG_DIR}" || return $?
  else
    local DISCOVERY_RELEASE_YAML=${TMP_DIR}/"discovery-${LATEST_RELEASE_VERSION}.yaml"
    # Download the latest release of Knative Discovery.
    wget "${knd_config}" -O "${DISCOVERY_RELEASE_YAML}" \
      || fail_test "Unable to download latest knative/discovery file."

    # Replace the default system namespace with the test's system namespace.
    sed -i "s/namespace: ${KNATIVE_DEFAULT_NAMESPACE}/namespace: ${TEST_DISCOVERY_NAMESPACE}/g" ${DISCOVERY_RELEASE_YAML}

    echo "Knative Discovery YAML: ${DISCOVERY_RELEASE_YAML}"

    kubectl apply -f "${DISCOVERY_RELEASE_YAML}" || return $?
    UNINSTALL_LIST+=( "${DISCOVERY_RELEASE_YAML}" )
  fi

  scale_controlplane webhook controller

  wait_until_pods_running ${TEST_DISCOVERY_NAMESPACE} || fail_test "Knative Discovery did not come up"

  echo "check the config map"
  kubectl get configmaps -n ${TEST_DISCOVERY_NAMESPACE}
}

function install_head {
  # Install Knative Discovery from HEAD in the current cluster.
  echo ">> Installing Knative Discovery from HEAD"
  install_knative_discovery || \
    fail_test "Knative HEAD installation failed"
}

function install_latest_release() {
  header ">> Installing Knative Discovery latest public release"
  local url="https://github.com/knative-sandbox/discovery/releases/download/${LATEST_RELEASE_VERSION}"
  local yaml="discovery.yaml"

  install_knative_discovery \
    "${url}/${yaml}" || \
    fail_test "Knative Discovery latest release installation failed"
}

function unleash_quacken() {
  echo "enable debug logging"
  cat test/config/config-logging.yaml | \
    sed "s/namespace: ${KNATIVE_DEFAULT_NAMESPACE}/namespace: ${TEST_DISCOVERY_NAMESPACE}/g" | \
    ko apply --strict -f - || return $?

  echo "unleash the quacken"
  cat test/config/quacken.yaml | \
    sed "s/namespace: ${KNATIVE_DEFAULT_NAMESPACE}/namespace: ${TEST_DISCOVERY_NAMESPACE}/g" | \
    ko apply --strict -f - || return $?
}

# Teardown the Knative environment after tests finish.
function knative_teardown() {
  echo ">> Stopping Knative Discovery"
  echo "Uninstalling Knative Discovery"
  ko delete --ignore-not-found=true --now --timeout 60s -f ${DISCOVERY_CONFIG}
  wait_until_object_does_not_exist namespaces ${TEST_DISCOVERY_NAMESPACE}

  echo ">> Uninstalling dependencies"
  for i in ${!UNINSTALL_LIST[@]}; do
    # We uninstall elements in the reverse of the order they were installed.
    local YAML="${UNINSTALL_LIST[$(( ${#array[@]} - $i ))]}"
    echo ">> Bringing down YAML: ${YAML}"
    kubectl delete --ignore-not-found=true -f "${YAML}" || return 1
  done
}

# Add function call to trap
# Parameters: $1 - Function to call
#             $2...$n - Signals for trap
function add_trap() {
  local cmd=$1
  shift
  for trap_signal in $@; do
    local current_trap="$(trap -p $trap_signal | cut -d\' -f2)"
    local new_cmd="($cmd)"
    [[ -n "${current_trap}" ]] && new_cmd="${current_trap};${new_cmd}"
    trap -- "${new_cmd}" $trap_signal
  done
}

# Setup resources common to all discovery tests.
function test_setup() {
  echo ">> Setting up logging..."

  # Install kail if needed.
  if ! which kail >/dev/null; then
    bash <(curl -sfL https://raw.githubusercontent.com/boz/kail/master/godownloader.sh) -b "$GOPATH/bin"
  fi

  # Capture all logs.
  kail >${ARTIFACTS}/k8s.log.txt &
  local kail_pid=$!
  # Clean up kail so it doesn't interfere with job shutting down
  add_trap "kill $kail_pid || true" EXIT

  install_test_resources || return 1

  echo ">> Publish test images"
  "$(dirname "$0")/upload-test-images.sh" e2e || fail_test "Error uploading test images"
}

function dump_extra_cluster_state() {
  # Collecting logs from all knative's discovery pods.
  echo "============================================================"
  local namespace=${TEST_DISCOVERY_NAMESPACE}
  for pod in $(kubectl get pod -n $namespace | grep Running | awk '{print $1}'); do
    for container in $(kubectl get pod "${pod}" -n $namespace -ojsonpath='{.spec.containers[*].name}'); do
      echo "Namespace, Pod, Container: ${namespace}, ${pod}, ${container}"
      kubectl logs -n $namespace "${pod}" -c "${container}" || true
      echo "----------------------------------------------------------"
      echo "Namespace, Pod, Container (Previous instance): ${namespace}, ${pod}, ${container}"
      kubectl logs -p -n $namespace "${pod}" -c "${container}" || true
      echo "============================================================"
    done
  done
}

function wait_for_file() {
  local file timeout waits
  file="$1"
  waits=300
  timeout=$waits

  echo "Waiting for existance of file: ${file}"

  while [ ! -f "${file}" ]; do
    # When the timeout is equal to zero, show an error and leave the loop.
    if [ "${timeout}" == 0 ]; then
      echo "ERROR: Timeout (${waits}s) while waiting for the file ${file}."
      return 1
    fi

    sleep 1

    # Decrease the timeout of one
    ((timeout--))
  done
  return 0
}
