#!/bin/bash

case $1 in
  postgres)
    chown -R postgres:postgres /data
    chmod 0700 /data
    shift
    exec sudo \
      -u postgres \
      -E -H \
      /bin/drycc-postgres $*
    ;;
  api)
    shift
    exec /bin/drycc-postgres-api $*
    ;;
  *)
    echo "Usage: $0 {postgres|api}"
    exit 2
    ;;
esac
