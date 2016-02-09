# -*- mode: ruby -*-
# vi: set ft=ruby :

#  For a complete reference, please see the online documentation at
# https://docs.vagrantup.com.

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "phusion-open-ubuntu-14.04-amd64"
  config.vm.box_url = "https://oss-binaries.phusionpassenger.com/vagrant/boxes/latest/ubuntu-14.04-amd64-vbox.box"

  config.vm.provider :vmware_fusion do |f, override|
    override.vm.box_url = "https://oss-binaries.phusionpassenger.com/vagrant/boxes/latest/ubuntu-14.04-amd64-vmwarefusion.box"
  end

  config.vm.host_name = "alkasir"

  config.vm.synced_folder ".", "/home/vagrant/src/github.com/alkasir/alkasir/"
  # configure portforwarding for the docker compose run databases
  config.vm.network "forwarded_port", guest: 39550, host: 39550
  config.vm.network "forwarded_port", guest: 39558, host: 39558
  config.vm.network "forwarded_port", guest: 8899, host: 8899

  # Only run the provisioning on the first 'vagrant up'
  if Dir.glob("#{File.dirname(__FILE__)}/.vagrant/machines/default/*/id").empty?
    config.vm.provision :shell, :path => "provision/base.sh"
    config.vm.provision :shell, :path => "provision/docker.sh"
    config.vm.provision :shell, :path => "provision/docker-vagrant.sh"
    config.vm.provision :shell, :path => "provision/docker-compose.sh"
    config.vm.provision :shell, :path => "provision/nodejs.sh"
    config.vm.provision :shell, :path => "provision/go.sh"
    config.vm.provision :shell, :path => "provision/maxminddb.sh", privileged: false
    config.vm.provision :shell, :path => "provision/env-vagrant.sh"
    config.vm.provision :shell, :path => "provision/profile-vagrant.sh", privileged: false
    # config.vm.provision :shell, :path => "provision/liquibase-update.sh"
  end
  # config.vm.provision :shell, :path => "provision/env-vagrant.sh"
  # config.vm.provision :shell, :path => "provision/docker-compose-up.sh", privileged: false

end
