# alkasir-torpt-server

## Generating a new obfs4 config

Start obfs4 proxy like this and all relevant data will be found in the
obfs4-state directory:

```sh
TOR_PT_MANAGED_TRANSPORT_VER=1 TOR_PT_SERVER_TRANSPORTS=obfs4 TOR_PT_SERVER_BINDADDR=obfs4-127.0.0.1:9999 TOR_PT_STATE_LOCATION=obfs4-state TOR_PT_ORPORT=127.0.0.1:9998 obfs4proxy 

```
