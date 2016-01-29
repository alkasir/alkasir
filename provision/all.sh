#!/bin/bash
# wrapper script for manual install outside vagrant. Privisions a development environment in a clean ubuntu inst
echo install all

set -e

p() {
  echo
  echo provisioning ${1}...
  echo
  (sudo bash provision/${1}.sh)
  echo '-------'
}

p env-local
p base
p docker
p docker-compose
p nodejs
p go
p maxminddb
sudo chown -R ${USER} ${HOME}/.alkasir-central
p bgpdump
