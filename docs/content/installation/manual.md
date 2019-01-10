---
title: Manual Installation
layout: docs
---

# Manual Installation

Drycc can be installed using our install script on **Ubuntu 16.04** and **14.04** amd64.

We recommend starting with a clean Ubuntu installation on machines with at least
2GB of RAM, 40GB of storage, and two CPU cores each. It's possible to run Drycc
on servers with lower specs, but we don't recommend it.

Before we get going with the installation, please note that if you plan on
running a multi-node cluster, you should boot at least 3 nodes to provide high
availability.

*Note: If you are installing on a provider that uses a customized kernel by
default, you may need to use the Ubuntu-supplied distribution kernel instead of
the custom kernel for ZFS filesystem support. On Linode, [use this
guide](https://www.linode.com/docs/tools-reference/custom-kernels-distros/run-a-distribution-supplied-kernel-with-kvm)
for instructions on how to switch.*

## Installation

Download and run the Drycc install script:

```
$ sudo bash < <(curl -fsSL https://dl.drycc.cc/install-drycc)
```

If you would rather take a look at the contents of the script before running it as root, download and
run it in two separate steps:

```
$ curl -fsSL -o /tmp/install-drycc https://dl.drycc.cc/install-drycc
... take a look at the contents of /tmp/install-drycc ...
$ sudo bash /tmp/install-drycc
```

_To install a [specific channel or version](https://releases.drycc.cc), you can
use the `--channel` and `--version` flags._

Running the installer script will:

1. Install Drycc's runtime dependencies
2. Download, verify and install the `drycc-host` binary
3. Download and verify filesystem images for each of the Drycc components
4. Install an Upstart job for controlling the `drycc-host` daemon

Some of the filesystem images are quite large (several hundred megabytes) so step 3 could take a while depending on
your Internet connection.

## Rinse and Repeat

You should install Drycc as above on every host that you want to be in the Drycc cluster.

## Set Up Nodes

First, ensure that all network traffic is allowed between all nodes in the cluster (specifically
all UDP and TCP packets). The following ports should also be open externally on the firewalls
for all nodes in the cluster:

* 80 (HTTP)
* 443 (HTTPS)
* 3000 to 3500 (user-defined TCP services, optional)

**Note:** A firewall with this configuration is _required_ to prevent external
access to internal management APIs.

The next step is to configure a Layer 0 cluster by starting the drycc-host
daemon on all nodes. The daemon uses Raft for leader election, and it needs to
be aware of all of the other nodes for it to function correctly.

If you are starting more than one node, the cluster should be configured using
a discovery token.  `drycc-host init` is a tool that handles generating and
configuring the token.

On the first node, create a new token with the `--init-discovery` flag. The
minimum multi-node cluster size is three, and this command does not need to be
run if you are only starting a single node.

```
$ sudo drycc-host init --init-discovery
https://discovery.drycc.cc/clusters/53e8402e-030f-4861-95ba-d5b5a91b5902
```

On the rest of the nodes, configure the generated discovery token:

```
$ sudo drycc-host init --discovery https://discovery.drycc.cc/clusters/53e8402e-030f-4861-95ba-d5b5a91b5902
```

## Start Drycc

Now, start the daemon and check that it has started:

**Ubuntu 14.04**

```
$ sudo start drycc-host
$ sudo status drycc-host
```

**Ubuntu 16.04**

```
$ sudo systemctl start drycc-host
$ sudo systemctl status drycc-host
```

If the status is `stop/waiting`, the daemon has failed to start. Check the log
file (`/var/log/drycc/drycc-host.log`) for any errors and try starting the
daemon again.

## Bootstrap Drycc

After you have running `drycc-host` instances, you can now bootstrap the cluster
with `drycc-host bootstrap`. You'll need a domain name with DNS A records
pointing to every node IP address and a second, wildcard domain CNAME to the
cluster domain.

**Example**

```
demo.local.drycc.cc.    A      192.168.84.42
demo.local.drycc.cc.    A      192.168.84.43
demo.local.drycc.cc.    A      192.168.84.44
*.demo.local.drycc.cc.  CNAME  demo.local.drycc.cc.
```

Set `CLUSTER_DOMAIN` to the main domain name and start the bootstrap process,
specifying the number of hosts that are expected to be present and the discovery
token if you created one.

```
$ sudo \
    CLUSTER_DOMAIN=demo.local.drycc.cc \
    drycc-host bootstrap \
    --min-hosts 3 \
    --discovery https://discovery.drycc.cc/clusters/53e8402e-030f-4861-95ba-d5b5a91b5902
```

*Note: You only need to run this on a single node in the cluster. It will
schedule jobs on nodes across the cluster as required.*

The bootstrapper will get all of necessary services running within Drycc. The
final log line will contain configuration that may be used with the
[command-line interface](/docs/cli).

If run into a problem while following these instructions, ensure that network
traffic is flowing unimpeded through the `flannel.1`, `dryccbr0`, and `veth*`
network interfaces and then open a GitHub issue describing the problem.

Now that you have Drycc installed and running, head over to the [Drycc
Basics](/docs/basics) page for a tutorial on deploying an application to Drycc.

## Release Mailing List

If you'd like to receive email about each month's stable release and security
updates, subscribe here:

<form action="https://drycc.us7.list-manage.com/subscribe/post?u=9600741fc187618e1baa39a58&id=8aadb709f3" method="post" target="_blank" novalidate class="mailing-list-form">
  <label>Email Address&nbsp;
    <input type="email" name="EMAIL" placeholder="you@example.com">
  </label>
  <button type="submit" name="subscribe">Subscribe</button>
</form>
