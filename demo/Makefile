
# Providers: virtualbox, vmware_fusion, aws
PROVIDER ?= virtualbox

# Domain to use when starting drycc
CLUSTER_DOMAIN ?= demo.local.drycc.cc
CLUSTER ?= demo-${PROVIDER}

# Location to install drycc CLI bin
CLI_INSTALL_BIN ?= /usr/local/bin/drycc

# ENV to set when running vagrant
# See https://github.com/mitchellh/vagrant/issues/4367 for use of SSH_AUTH_SOCK=''
VAGRANT_ENV := CLUSTER_DOMAIN="${CLUSTER_DOMAIN}" SSH_AUTH_SOCK="" FLAG_BOOTSTRAP=false


.PHONY: init bootstrap up down halt update ssh reset token destroy dashboard ssh-cmd


# Initialize a new cluster
init: update
	@$(MAKE) up
	@$(MAKE) bootstrap
	@printf "\nDrycc is now running in the VM.\nOpen dashboard: make dashboard\n\n"

# Install drycc cli on host
$(CLI_INSTALL_BIN):
	curl -sL -A "`uname -sp`" https://dl.drycc.cc/cli | zcat >${CLI_INSTALL_BIN}
	chmod +x ${CLI_INSTALL_BIN}

# Bootstrap the cluster
# This should only be called once per cluster
# Runs bootstrap in VM and cluster add on host
bootstrap: $(CLI_INSTALL_BIN)
	@$(MAKE) CMD='CLUSTER_DOMAIN="${CLUSTER_DOMAIN}" drycc-host bootstrap /etc/drycc/bootstrap-manifest.json 2>&1' ssh-cmd | tee bootstrap.log
	@(grep '^drycc cluster add .*$$' bootstrap.log || echo exit 1) | sed "s/ default / ${CLUSTER} /" | sed "s/$$(printf '\r')\$$/ --force --default/" | bash
	@rm bootstrap.log

# Start the VM and drycc
up:
	${VAGRANT_ENV} vagrant up --provider ${PROVIDER}
	@$(MAKE) CMD='sudo systemctl start drycc-host || echo drycc-host already running or unable to start' ssh-cmd

# Stop the VM
down:
	${VAGRANT_ENV} vagrant halt

halt: down

# Download the latest drycc VM image
# This does not update already initialized VMs
update:
	${VAGRANT_ENV} vagrant box update

# SSH into VM
ssh:
	${VAGRANT_ENV} vagrant ssh

# Destroy and rebuilt the VM
reset:
	@$(MAKE) destroy
	@$(MAKE) update
	@$(MAKE) init

# Destroy VM
# Run this to delete the VM and free up disk space
destroy: $(CLI_INSTALL_BIN)
	${VAGRANT_ENV} vagrant destroy -f
	drycc cluster remove ${CLUSTER}

# Get token for API and dashboard
token: $(CLI_INSTALL_BIN)
	@drycc -a dashboard -c ${CLUSTER} env get LOGIN_TOKEN

# Open dashboard in browser
dashboard:
	@CMD=open && which xdg-open > /dev/null && CMD=xdg-open; \
	$$CMD http://dashboard.${CLUSTER_DOMAIN}?token=${$(MAKE) token}

# Run a command in the VM via SSH
ssh-cmd:
	${VAGRANT_ENV} vagrant ssh -c '${CMD}'

