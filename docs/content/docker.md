---
title: Docker
layout: docs
toc_min_level: 2
---

# Docker

Drycc has a built-in `docker-receive` app which wraps a Docker registry and
imports pushed Docker images into a Drycc cluster.

## Configuration

Before pushing images to a Drycc cluster, both the local `drycc` and `docker`
CLIs need to be configured.

Configure the `drycc` CLI by running:

```
$ drycc docker set-push-url
```

Configure the `docker` CLI by running:

```
$ drycc docker login
```

_If you see "Error configuring docker", follow the instructions which appear
above the error then re-run `drycc docker login`._

## Push an image

Run the following to push a Docker image to `docker-receive` and deploy it:

```
$ drycc -a APPNAME docker push IMAGE
```

where `APPNAME` is the name of an existing Drycc app and `IMAGE` is a reference
to a Docker image which is available to the local `docker` CLI (in other words,
an image which appears in the output of `docker images`).

## Routing

Drycc automatically registers the HTTP route `http://APPNAME.$CLUSTER_DOMAIN`
for the app. In order to receive HTTP traffic for this route, the app needs to
listen on the port which is set in the `PORT` environment variable.

## Example

Here is an example of deploying the [Drycc Node.js example app](https://github.com/drycc-examples/nodejs-drycc-example)
using `drycc docker push`.

Clone the git repository:

```
$ git clone https://github.com/drycc-examples/nodejs-drycc-example.git
$ cd nodejs-drycc-example
```

Build the Docker image:

```
$ docker build -t nodejs-drycc-example .
```

Create an app:

```
$ drycc create --remote "" nodejs
Created nodejs
```

_NOTE: the `--remote ""` flag prevents Drycc trying to configure the local
git repository with a `drycc` remote, something which is useful only when
deploying with `git push` rather than `drycc docker push`._

Push the Docker image:

```
$ drycc -a nodejs docker push nodejs-drycc-example
drycc: getting image config with "docker inspect -f {{ json .Config }} nodejs-drycc-example"
drycc: tagging Docker image with "docker tag nodejs-drycc-example docker.1.localdrycc.com/nodejs:latest"
drycc: pushing Docker image with "docker push docker.1.localdrycc.com/nodejs:latest"
The push refers to a repository [docker.1.localdrycc.com/nodejs] (len: 1)
82b9b0ffb6da: Pushed
be8edf33c031: Pushed
...
767584930cea: Pushed
d34921bc2709: Pushed
latest: digest: sha256:be6aeade058f0df30a039a432aaf4cb21accd992d4c0df80ddb333b15f401b6a size: 16114
drycc: image pushed, waiting for artifact creation
drycc: deploying release using artifact URI http://drycc:dbd202007171356f4551160dede351ae@docker-receive.discoverd?name=nodejs&id=sha256:be6aeade058f0df30a039a432aaf4cb21accd992d4c0df80ddb333b15f401b6a
drycc: image deployed, scale it with 'drycc scale app=N'

```

You now have a release with an `app` process which runs using the pushed image
and has an `ENTRYPOINT`, `CMD` and `ENV` as taken from the Docker image's
config.

If this is the first deploy of the app, scale the `app` process to start it:

```
$ drycc -a nodejs scale app=1
scaling app: 0=>1

16:24:31.397 ==> app 4a7319af-af2c-4fe1-9a9a-2dd4d5bd3765 pending
16:24:31.398 ==> app host0-4a7319af-af2c-4fe1-9a9a-2dd4d5bd3765 starting
16:24:31.527 ==> app host0-4a7319af-af2c-4fe1-9a9a-2dd4d5bd3765 up

scale completed in 140.63784ms
```

The `app` process will be configured with a service name like `APPNAME-web` so
your Drycc apps can communicate with the deployed service internally using
`APPNAME-web.discoverd:PORT` (e.g. `nodejs-web.discoverd:8080`):

```
$ drycc -a nodejs run curl http://nodejs-web.discoverd:8080
Hello from Drycc on port 8080 from container 4a7319af-af2c-4fe1-9a9a-2dd4d5bd3765
```

The app can be reached externally via the automatically registered route
`http://APPNAME.$CLUSTER_DOMAIN`:

```
$ curl http://nodejs.1.localdrycc.com
Hello from Drycc on port 8080 from container 4a7319af-af2c-4fe1-9a9a-2dd4d5bd3765
```
