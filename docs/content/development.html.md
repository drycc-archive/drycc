---
title: Development
layout: docs
toc_min_level: 2
---

# Development

This guide will explain how to:

* Make changes to the Drycc source code
* Build and run Drycc
* Run the tests
* Create a release of Drycc

## Development environment

Development work is typically done inside a [VirtualBox](https://www.virtualbox.org/) VM managed
by [Vagrant](https://www.vagrantup.com/), and Drycc includes a Vagrantfile which fully automates
the creation of the VM.

### Running the development VM

If you don't already have VirtualBox and Vagrant installed, you should
install them by following the directions on their respective web sites.

Clone the Drycc source code locally:

```
$ git clone https://github.com/drycc/drycc.git
```

Then, inside the `drycc` directory, bring up the VM:

```
$ vagrant up
```

If this is the first time you are creating the VM, Vagrant will need to download the
underlying VirtualBox files which are ~1GB in size, so this could take a while depending on
your internet connection.

Once that command has finished, you will have a VM running which you can SSH to:

```
$ vagrant ssh
```

From now on, it is assumed that commands will be run inside the VM, unless otherwise stated.

## Making code changes

The development VM is configured to share the Drycc source code from your machine and mount
it inside the VM, meaning you can edit files locally on your machine and those changes
will be visible inside the VM.

Since Drycc is primarily written in Go, the source code needs to be inside a valid Go workspace.
The development VM has a `GOPATH` of `$HOME/go` and the Drycc source code is synchronized into
`$GOPATH/src/github.com/drycc/drycc`.

If you don't have a specific issue you are trying to fix, but are interested in contributing
to the project, you should start by looking at GitHub issues labelled
[complexity/easy](https://github.com/drycc/drycc/labels/complexity/easy).

## Building Drycc

We use the [tup](http://gittup.org/tup/) build system to run the commands which build the various
components of Drycc.

To kickoff the build process, just run `make`:

```
$ make
```

This will do things like build Go binaries and create Docker images. If you're interested in
exactly what will be built, take a look at the `Tupfiles` in various subdirectories.

If any build command fails, `tup` will output an error and abort the entire build. You can then
fix the issue and then re-run `make`.

If you want to rebuild all Go binaries, run `make clean`.

Once tup runs successfully, you will have a number of built Go binaries and Docker images which
can be used to run Drycc.

## Running Drycc

Once you have built all the Drycc components, you can boot a single node Drycc cluster by running
the following script:

```
$ script/bootstrap-drycc
```

This will do the following:

* stop the `drycc-host` daemon and any running Drycc services
* start the `drycc-host` daemon
* run the Drycc bootstrapper, which will start all of the Drycc services

If you want to boot Drycc using a different job backend, or an external IP other
than that of the `eth0` device, the script provides some options for doing so.
See `script/bootstrap-drycc -h` for a full list of supported options.

Once Drycc is running, you can add the cluster to the `drycc` CLI tool using the
bootstrap output, and then try out your changes.

## Debugging

If things don't seem to be running as expected, here are some useful commands to help
diagnose the issue:

### check the drycc-host daemon log

```
$ less /tmp/drycc-host.log
```

### view running jobs

```
$ drycc-host ps
ID                                                      STATE    STARTED             CONTROLLER APP  CONTROLLER TYPE
drycc-66f3ca0c60374a1abb172e3a73b50e21                  running  About a minute ago  example         web
drycc-9d716860f69f4f63bfb4074bcc7f4419                  running  4 minutes ago       gitreceive      app
drycc-7eff6d37af3c4d909565ca0ab3b077ad                  running  4 minutes ago       router          app
drycc-b8f3ecd48bb343dab96744a17c96b95d                  running  4 minutes ago       blobstore       web
...
```

### view all jobs (i.e. running + stopped)

```
$ drycc-host ps -a
ID                                                      STATE    STARTED             CONTROLLER APP  CONTROLLER TYPE
drycc-7fd8c48542e442349c0217e7cb52dec9                  running  15 seconds ago      example         web
drycc-66f3ca0c60374a1abb172e3a73b50e21                  running  About a minute ago  example         web
drycc-9868a539703145a1886bc2557f6f6441                  done     2 minutes ago       example         web
drycc-100f36e9d18849658e11188a8b85e79f                  done     3 minutes ago       example         web
drycc-f737b5ece2694f81b3d5efdc2cb8dc56                  done     4 minutes ago
drycc-9d716860f69f4f63bfb4074bcc7f4419                  running  5 minutes ago       gitreceive      app
drycc-7eff6d37af3c4d909565ca0ab3b077ad                  running  5 minutes ago       router          app
drycc-b8f3ecd48bb343dab96744a17c96b95d                  running  5 minutes ago       blobstore       web
...
```

### view the output of a job

```
$ drycc-host log $JOBID
Listening on 55006
```

### inspect a job

```
$ drycc-host inspect $JOBID
ID                           drycc-075c21e8b79a41d89713352f04f94a71
Status                       running
StartedAt                    2014-10-14 14:34:11.726864147 +0000 UTC
EndedAt                      0001-01-01 00:00:00 +0000 UTC
ExitStatus                   0
IP Address                   192.168.200.24
drycc-controller.release     954b7ee40ef24a1798807499d5eb8297
drycc-controller.type        web
drycc-controller.app         e568286366d443c49dc18e7a99f40fc1
drycc-controller.app_name    example
```

### stop a job

```
$ drycc-host stop $JOBID
```

### stop all jobs

```
$ drycc-host ps -a | xargs drycc-host stop
```

*(NOTE: as jobs are stopped, the scheduler may start new jobs. To avoid this, stop the scheduler first)*

### stop all jobs for a particular app

Assuming the app has name `example`:

```
$ drycc-host ps | awk -F " {2,}" '$4=="example" {print $1}' | xargs drycc-host stop
```

### upload logs and system information to a GitHub gist

If you want to get help diagnosing issues on your system, run the following to upload some
useful information to an anonymous GitHub gist:

```
$ drycc-host collect-debug-info
INFO[03-11|19:25:29] uploading logs and debug information to a private, anonymous gist
INFO[03-11|19:25:29] this may take a while depending on the size of your logs
INFO[03-11|19:25:29] getting drycc-host logs
INFO[03-11|19:25:29] getting job logs
INFO[03-11|19:25:29] getting system information
INFO[03-11|19:25:30] creating anonymous gist
789.50 KB / 789.50 KB [=======================================================] 100.00 % 93.39 KB/s 8s
INFO[03-11|19:25:38] debug information uploaded to: https://gist.github.com/47379bd4604442cac820
```

You can then post the gist in the `#drycc` IRC room when asking for assistance to make it easier for
someone to help you.

If you would rather not use the GitHub gist service, or your logs are too big to fit into a single gist,
you can create a tarball of the information by specifying the `--tarball` flag:

```
$ drycc-host collect-debug-info --tarball
INFO[03-11|19:28:58] creating a tarball containing logs and debug information
INFO[03-11|19:28:58] this may take a while depending on the size of your logs
INFO[03-11|19:28:58] getting drycc-host logs
INFO[03-11|19:28:58] getting job logs
INFO[03-11|19:28:58] getting system information
INFO[03-11|19:28:59] created tarball containing debug information at /tmp/drycc-host-debug407848418/drycc-host-debug.tar.gz
```

You can then send this to a Drycc developer after speaking to them in IRC.

## Running tests

Drycc has two types of tests:

* "unit" tests which are run using `go test`
* "integration" tests which run against a booted Drycc cluster

### Run the unit tests

To run all the unit tests:


```
$ go test ./...
```

To run tests for an individual component (e.g. the router):

```
$ go test ./router
```

To run tests for a component and all sub-components (e.g. the controller):

```
$ go test ./controller/...
```

### Run the integration tests

The integration tests live in the `tests` directory, and require a running Drycc
cluster before they can run.

To run all the integration tests:

```
$ script/run-integration-tests
```

This will:

* Run `make` to build Drycc
* Boot a single node Drycc cluster by running `script/bootstrap-drycc`
* Run the integration test binary (i.e. `bin/drycc-test`)

To run an individual integration test (e.g. `TestEnvDir`):

```
$ script/run-integration-tests -f TestEnvDir
```

## Pull request

Once you have made changes to the Drycc source code and tested your changes, you
should open a pull request on GitHub so we can review your changes and merge
them into the Drycc repository.

Please see the [contribution guide](/docs/contributing) for more information.

## Releasing Drycc

Once you have built and tested Drycc inside the development VM, you can create a release
and install the components on other hosts (e.g. in EC2).

A Drycc release is a set of components consisting of binaries, configuration files and
filesystem images, all of which must be installed in order to run Drycc.

### The Update Framework (TUF)

Drycc uses [The Update Framework](http://theupdateframework.com/) (also known as TUF) to
securely distribute all of the Drycc components.

To release Drycc, you will need to generate a TUF repository using the
[go-tuf](https://github.com/drycc/go-tuf) library.

Follow the [installation instructions](https://github.com/drycc/go-tuf#install)
and then follow the ["Create signed root manifest"](https://github.com/drycc/go-tuf#examples)
example. You should now have a directory with the following layout:


```
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
└── staged
    ├── root.json
    └── targets
```

Commit the repository by running the following:

```
$ tuf add
$ tuf snapshot
$ tuf timestamp
$ tuf commit
```

You should now see some files in the `repository` directory (the filenames may differ):

```
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── 966421c866be59e54fca0d6e2c0f62f1790934a25c065715dce0fa3378b7200a1a47b18d60647db115f6a2c8f9bd90d96985c15799c0429f985704446c39a9d3.targets.json
│   ├── a0f8878b84aa629271c614b6816f9e9fd12dc2d880ccdf03b3141b09b37d5b16c92e9f84c21fd064aff57b90f31c9711d4c84b5a22e70e03a21943511a905cd4.root.json
│   ├── a8a2ce2842ffab18a5217c6192735915ed3685baad05c25b6dd499634abefaa1b62a86de2640b6bedbdc21add582d8bd01ca615b740c26c019058414c8f887eb.snapshot.json
│   ├── root.json
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
└── staged
    └── targets
```

Upload the `repository` directory to a location where you intend to serve the
released version of Drycc from. For example, if using the `my-drycc-repo` S3
bucket:

```
$ aws s3 cp --recursive --acl public-read repository s3://my-drycc-repo/tuf
```

### Rebuild Drycc

The TUF root keys need to be compiled into the Drycc release, and image URLs must
be relative to the file server the TUF repository is served from. This can be
accomplished by updating the `tuf` object in `builder/manifest.json` then re-running
`make`, e.g.:

```
"tuf": {
  "repository": "https://s3.amazonaws.com/my-drycc-repo/tuf",
  "root_keys": [
    {"keytype":"ed25519","keyval":{"public":"31351ecc833417968faabf98e004d1ef48ecfd996f971aeed399a7dc735d2c8c"}}
  ]
}
```

*The TUF root keys can be determined by running `tuf root-keys` in the TUF repository.*

Run `script/build-drycc --version XXX` instead of `make` to set an explicit version:

```
$ script/build-drycc --version v20171206.lmars
```

### Export components

Export the Drycc components into the TUF repository (it expects the targets, snapshot and
timestamp passphrases to be set in the `TUF_TARGETS_PASSPHRASE`, `TUF_SNAPSHOT_PASSPHRASE`
and `TUF_TIMESTAMP_PASSPHRASE` environment variables respectively):

```
$ script/export-components /path/to/tuf-repo
```

### Upload components

Upload the `repository` directory of the TUF repo to the remote file server.

For example, if using the `my-drycc-repo` S3 bucket:

```
$ aws s3 cp --recursive --acl public-read repository s3://my-drycc-repo/tuf
```

You can now distribute `script/install-drycc` and run it with an explicit repo URL
to install the custom built Drycc components:

```
install-drycc -r https://s3.amazonaws.com/my-drycc-repo
```
