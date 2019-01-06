---
title: Redis
layout: docs
---

# Redis

The Drycc Redis appliance provides Redis 3.0 in a single process configuration.
The data stored in this process is ephemeral and is intended for caching and
development use.

## Usage

### Adding a server to an app

Redis comes ready to go as soon as you've installed Drycc. After you create
an app, you can provision a server for your app by running:

```text
drycc resource add redis
```

This will provision a Redis server as a Drycc app and configure your application
to connect to it.

### Connecting to the database

Provisioning the database will add a few environment variables to your app
release. `REDIS_HOST`, `REDIS_PORT`, and `REDIS_PASSWORD` provide connection
details for the database.

Drycc will also create the `REDIS_URL` environment variable which is utilized
by some libraries to configure connections.

### Connecting to a console

To connect to a console for the database, run `drycc redis redis-cli`. This does
not require the Redis client to be installed locally or firewall/security
changes, as it runs in a container on the Drycc cluster.

### External access

An external route can be created that allows access to the database from
services that are not running on Drycc.

```text
drycc -a $(drycc env get DRYCC_REDIS) route add tcp --service $(drycc env get DRYCC_REDIS) --leader
```

This will provision a TCP port that always points at the primary instance.

For security reasons this port should be firewalled, and it should only be
accessed over the local network, VPN, or SSH tunnel.

## Safety

No safety or availability guarantees are currently provided for the Redis
appliance. Data loss and inconsistency is likely. Any data stored should be
treated as ephemeral and only used for caching, development, and testing.
