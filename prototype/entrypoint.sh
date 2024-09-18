#!/bin/sh
set -e

# If first run, run all build scripts
if [ ! -f "/run/solo-build.lock" ]; then
  cp -R /build-scripts /.build-scripts.tmp
  chmod a+x /.build-scripts.tmp/*

  for file in /.build-scripts.tmp/*; do
    if [ -f "$file" ]; then
      . "$file"
    fi
  done

  rm -rf /.build-scripts.tmp
  touch "/run/solo-build.lock"
fi

# Run all run scripts
cp -R /run-scripts /.run-scripts.tmp
chmod a+x /.run-scripts.tmp/*

for file in /.run-scripts.tmp/*; do
  if [ -f "$file" ]; then
    . "$file"
  fi
done

rm -rf /.run-scripts.tmp

# todo: allow switching users at this point

# Run container command using exec
exec "$@"
