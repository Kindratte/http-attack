#!/bin/sh

FREQUENCY=${FREQUENCY:-"1000"}
SERVER_HOST=${SERVER_HOST:-"localhost"}
MINUTES=${MINUTES:-"1"}
NUM_OF_CONNECTIONS=${NUM_OF_CONNECTIONS:-"10000"}
LOCATION=${LOCATION:-"2"}
LOGIN=${LOGIN:-""}
PASSWORD=${PASSWORD:-""}
TEST=${TEST:-true}
NO_AUTH=${NO_AUTH:-false}
CAS=${CAS:-false}

exec /http-attack/http-attack -f="${FREQUENCY}" -h="${SERVER_HOST}" -m="${MINUTES}" -c="${NUM_OF_CONNECTIONS}" -loc="$LOCATION" -l="$LOGIN" -p="$PASSWORD" -t="$TEST" -na="$NO_AUTH" -tc="$CAS"
