#!/bin/bash

gosu postgres postgres --single <<- EOSQL
    CREATE USER alkasir_central WITH PASSWORD 'alkasir_central';
    CREATE DATABASE alkasir_central;
    GRANT ALL PRIVILEGES ON DATABASE alkasir_central TO alkasir_central;
    ALTER USER alkasir_central WITH SUPERUSER;
EOSQL
