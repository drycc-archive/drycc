#!/bin/sh

case $1 in
  controller) exec /bin/drycc-controller ;;
  scheduler)  exec /bin/drycc-scheduler ;;
  worker)  exec /bin/drycc-worker ;;
  *)
    echo "Usage: $0 {controller|scheduler|worker}"
    exit 2
    ;;
esac
