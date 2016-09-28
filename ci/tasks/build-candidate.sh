#!/usr/bin/env bash

set -e

source gcp-tools-release/ci/tasks/utils.sh
source /etc/profile.d/chruby-with-ruby-2.1.2.sh

cpi_release_name="bosh-gcp-tools"
semver=`cat version-semver/number`

pushd gcp-tools-release
  echo "Using BOSH CLI version..."
  bosh version

  echo "Creating CPI BOSH Release..."
  bosh create release --name ${cpi_release_name} --version ${semver} --with-tarball
popd

image_path=gcp-tools-release/dev_releases/${cpi_release_name}/${cpi_release_name}-${semver}.tgz
echo -n $(sha1sum $image_path | awk '{print $1}') > $image_path.sha1

mv ${image_path} candidate/
mv ${image_path}.sha1 candidate/
