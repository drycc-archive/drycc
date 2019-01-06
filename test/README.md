# drycc-test

drycc-test contains full-stack acceptance tests for Drycc.

## Usage

### Bootstrap Drycc

The tests need a running Drycc cluster, so you will need to boot one first.

To run Drycc locally, first boot and SSH to the Drycc dev box:

```text
vagrant up
vagrant ssh
```

then build and bootstrap Drycc (this may take a few minutes):

```text
make
script/bootstrap-drycc
```

### Run the tests

Run the `drycc cluster add` command from the bootstrap output to add the cluster to your `~/.dryccrc` file, then run the tests:

```text
drycc cluster add ...
cd ~/go/src/github.com/drycc/drycc/test
bin/drycc-test --dryccrc ~/.dryccrc --cli `pwd`/../cli/bin/drycc
```

## Auto booting clusters

The test binary is capable of booting its own cluster to run the tests against, provided you are using a machine capable of running KVM.

### Build root filesystem + kernel

Before running the tests, you need a root filesystem and a Linux kernel capable of building and running Drycc.

To build these into `/tmp/drycc`:

```text
mkdir -p /tmp/drycc
sudo rootfs/build.sh /tmp/drycc
```

You should now have `/tmp/drycc/rootfs.img` and `/tmp/drycc/vmlinuz`.

### Build the tests

```text
go build -o drycc-test
```

### Download Drycc CLI

The tests interact with the VM cluster using the Drycc CLI, so you will need it locally.

Download it into the current directory:

```text
curl -sL -A "`uname -sp`" https://dl.drycc.cc/cli | zcat >drycc
chmod +x drycc
```

### Run the tests

```text
sudo ./drycc-test \
  --user `whoami` \
  --rootfs /tmp/drycc/rootfs.img \
  --kernel /tmp/drycc/vmlinuz \
  --cli `pwd`/drycc
```

## CI

### Install the runner

Install Git if not already installed, then check out the Drycc git repo and run
the following to install the runner into `/opt/drycc-test`:

```
sudo test/scripts/install
```

Add the following credentials to `/opt/drycc-test/.credentials`:

```
export AUTH_KEY=XXXXXXXXXX
export GITHUB_TOKEN=XXXXXXXXXX
export AWS_ACCESS_KEY_ID=XXXXXXXXXX
export AWS_SECRET_ACCESS_KEY=XXXXXXXXXX
```

Now start the runner:

```
sudo start drycc-test
```

### Updating the runner

If the runner code has been changed, restart the Upstart job to pull in the new changes:

```
sudo restart drycc-test
```

If the rootfs needs rebuilding, you will need to remove the existing image before starting
the runner again:

```
sudo stop drycc-test
sudo rm -rf /opt/drycc-test/build/{rootfs.img,vmlinuz}
sudo start drycc-test
```
