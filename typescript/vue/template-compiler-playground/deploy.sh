#!/usr/bin/env bash
set -e

yarn run build

cd ./dist

git init
git config user.name 'lx1036'
git config user.email 'lx1036@126.com'
git add -A
git commit -m 'deploy'

git push -f git@github.com:lx1036/vue-template-compiler-playground.git master:gh-pages

cd -
