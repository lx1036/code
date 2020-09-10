#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "kong" --dbname "konga" <<-EOSQL
    CREATE DATABASE konga;
    GRANT ALL PRIVILEGES ON DATABASE konga TO kong;
EOSQL
