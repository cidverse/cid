#! /usr/bin/env nix-shell
#! nix-shell -i bash -p git-cliff

# generate changelog
git cliff --context --output pkg/core/changelog/gocliff.json
