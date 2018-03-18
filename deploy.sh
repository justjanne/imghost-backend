#!/bin/sh
IMAGE=k8r.eu/justjanne/imghost
TAGS=$(git describe --always --tags HEAD)
DEPLOYMENT=imghost-backend
POD=imghost-backend

kubectl -n imghost set image deployment/$DEPLOYMENT $POD=$IMAGE:$TAGS