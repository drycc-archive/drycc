# create a user with a well known name / uid / gid
export USER="drycc"
export USER_UID="5000"
export USER_GID="5000"

groupadd \
  --gid "${USER_GID}" \
  "${USER}"

useradd \
  --uid     "${USER_UID}" \
  --gid     "${USER_GID}" \
  --comment "Drycc slug user" \
  --home    "/app" \
  "${USER}"
