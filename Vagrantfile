# Vagrantfile for myruntime development.
# Use this if you're developing on macOS or Windows and need a Linux VM.
#
# Usage:
#   vagrant up        — create and provision the VM
#   vagrant ssh       — SSH into the VM
#   vagrant halt      — stop the VM
#   vagrant destroy   — delete the VM
#
# The project directory is synced to /home/vagrant/myruntime inside the VM.

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/jammy64"    # Ubuntu 22.04 LTS
  config.vm.hostname = "myruntime-dev"

  # Private network for container networking tests
  config.vm.network "private_network", type: "dhcp"

  # Forward common test ports
  config.vm.network "forwarded_port", guest: 8080, host: 8080

  config.vm.provider "virtualbox" do |vb|
    vb.memory = "4096"
    vb.cpus = 2
    vb.name = "myruntime-dev"
  end

  # Provision: install Go, tools, and verify cgroup v2
  config.vm.provision "shell", inline: <<-SHELL
    set -e

    # Install Go
    GO_VERSION="1.22.0"
    if [ ! -f /usr/local/go/bin/go ]; then
      curl -sL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" | tar -C /usr/local -xzf -
    fi
    echo 'export PATH=$PATH:/usr/local/go/bin:~/go/bin' >> /home/vagrant/.bashrc

    # Install required packages
    apt-get update -qq
    apt-get install -y -qq build-essential iptables iproute2 bridge-utils \
      uidmap curl git tree

    # Verify cgroup v2
    if [ -f /sys/fs/cgroup/cgroup.controllers ]; then
      echo "cgroup v2: OK"
    else
      echo "WARNING: cgroup v2 not available in this VM"
      echo "You may need to add systemd.unified_cgroup_hierarchy=1 to kernel params"
    fi

    echo ""
    echo "=== VM ready ==="
    echo "Run: cd /home/vagrant/myruntime && make build"
  SHELL

  # Sync project directory
  config.vm.synced_folder ".", "/home/vagrant/myruntime"
end
