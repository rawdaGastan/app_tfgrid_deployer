#!/bin/sh

set -ex 

if [ -z ${REPO_NAME+x} ]
then
    echo 'Error! $REPO_NAME is required.'
    exit 64
fi

if [ -z ${BACKEND_DIR+x} ]
then
    echo 'Error! $BACKEND_DIR is required.'
    exit 64
fi

if [ -z ${FRONTEND_DIR+x} ]
then
    echo 'Error! $FRONTEND_DIR is required.'
    exit 64
fi

cd /mydata/$REPO_NAME
git pull
cd $BACKEND_DIR
docker build --force-rm -t $BACKEND_DIR .
cd ../$FRONTEND_DIR
docker build --force-rm -t $FRONTEND_DIR .
zinit stop $REPO_NAME
zinit start $REPO_NAME
