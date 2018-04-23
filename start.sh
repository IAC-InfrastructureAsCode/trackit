#!/bin/bash

set -ex

rm -f ./server/main
if [[ ! -f "server/main" ]]
then
	pushd "server"
	./buildstatic.sh
	popd
fi

pushd scripts
./copy_schema.sh
popd

pushd docker
docker-compose build
../scripts/awsenv default docker-compose up
popd
