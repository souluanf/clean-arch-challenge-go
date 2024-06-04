#!/bin/sh

set -e

host="$1"
shift
cmd="$@"

until nc -z "$host" 3306; do
  echo "Waiting for mysql at $host:3306..."
  sleep 1
done

exec $cmd