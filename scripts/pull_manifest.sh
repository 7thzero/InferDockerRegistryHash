#!/bin/bash
#
# Simplified example in shell script for how to pull a manifest from a docker image in dockerhub

#
# Constants
registryBase='https://registry-1.docker.io'
authBase='https://auth.docker.io'
authService='registry.docker.io'

#
# Approved/official images live in the 'library/$image' path
image="library/redis"

#
# Get a token to use to query the dockerhub API
#	If you are targeting a private or internal docker registry, you may have basic auth or other authentication to account for!
token="$(curl -fsSL "$authBase/token?service=$authService&scope=repository:$image:pull" | jq --raw-output '.token')"
echo "$token"

#
# Pull the manifest
curl -fsSL -H "Authorization: Bearer $token" -H 'Accept: application/vnd.docker.distribution.manifest.v2+json' "$registryBase/v2/$image/manifests/buster"
