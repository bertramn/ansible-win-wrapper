# Ansible Wrapper for Windows

This is a simple executable wrapper to allow Vagrant to run ansible commands on windows through to a configured cygwin environment. There are a few [threads](http://stackoverflow.com/questions/29743491/how-to-install-ansible-playbook-on-windows-as-host-with-vagrant) out there that document the use of batch files to bootstrap ansible but I found the windows command shell does not actually work properly when using `--extra-vars` for instance.

We assume you have installed a working cygwin shell on your workstation.

### Installation

##### Install Cygwin Packages

Use the cygwin package installer from https://cygwin.com/ and install the following packages:
* python
* python-setuptools
* gcc-g++
* wget
* openssh
* curl
* git

Now you should have base python environment installed in your cygwin environment.

##### Configure Python Environment

First lets open up a cygwin terminal and install pip.

```sh
$ easy_install-2.7 pip
```

Thereafter we will update `pip` and `setuptools`

```sh
$ pip install -U pip setuptools
```

##### Install Ansible

Still in the cygwin terminal we need to install the ansible dependencies.

```sh
$ pip install -U crypto paramiko PyYAML Jinja2 httplib2 six
```

And finally lets install ansible itself.

```sh
$ pip install -U ansible
```

Quick test to ensure it works.

```sh
$ ansible --version
ansible 2.0.0.2
  config file =
  configured module search path = Default w/o overrides
```

### Configure Vagrant Environment to work with Ansible

To bridge Vagrant on windows to Ansible on Cygwin we can now use the Ansible Wrapper for Windows binary. The Vagrant Ansible provisionser will use the Vagrant windows environment to load the `ansible-playbook` executable. So we need to ensure Vagrant will find the wrapper executable to hand off to the Cygwin envionment.

Simply copy the `ansible-win-wrapper.exe` into the same folder of the `vagrant.exe` and rename it to `ansible-playbook.exe`.

```cmd
copy ansible-win-wrapper C:\HashiCorp\Vagrant\bin\ansible-playbook.exe

vagrant --version
Vagrant 1.8.1

ansible-playbook --version
ansible-playbook 2.0.0.2
  config file =
  configured module search path = Default w/o overrides
```

Et voilà, ansible-playbook should now work from vagrant.

### Test Example

Setup a Vagrant project folder with a few basic files. I use cygwin terminal for simplicity but you can also do all of this in windows command or powershell.

```sh
$ mkdir ~/vagrant-test
$ cd ~/vagrant-test
```

All 4 files listed below go into the project folder. 

**`Vagrantfile`**

A basic Vagrant machine definition with ansible provisioner configuration.

```rb
# -*- mode: ruby -*-
# vi: set ft=ruby :
Vagrant.configure(2) do |config|

  config.vm.box = "hashicorp/precise64"

  config.vm .provision "ansible" do |ansible|
    ansible.verbose = "v"
    ansible.limit = "all"
    ansible.inventory_path = "./inventory"
    ansible.playbook = "playbook.yml"
  end

end
```

**`ansible.cnf`**

Ansible default configuration with some Windows specific config.

```properties
[defaults]
host_key_ckecking = False
[ssh_connection]
# ControlMaster on cygwin OpenSSH does not work, so disable it
control_path = none
```

**`inventory`**

Need to create an ansible inventory definition, the default Vagrant generated one does not work, vagrants ansible provisioner generates a inventory path arg that is Windows but must POSIX for ansible to work properly, so we just cheat a little.

```properties
default ansible_ssh_host=127.0.0.1 ansible_ssh_port=2222 ansible_ssh_user='vagrant' ansible_ssh_private_key_file='./.vagrant/machines/default/virtualbox/private_key'
```
Note: You may need to adjust the `ansible_ssh_port` to whatever the Vagrant machine bound to. You can get around all this by using a DNS plugin like `vagrant-hostmanager` and then just define actual DNS resolvable vagrant machine names in the inventory.

Also may need to `$ chmod 0644 inventory` depending on how your cygwin fstab is setup, ansible may complain about +x permissions.

**`playbook.yml`**

```yaml
---
- hosts: all
  tasks:
  - debug: "msg='Hello Ansible from Vagrant'"
```

Lets run it:

```sh
$ vagrant up --no-provision

Bringing machine 'default' up with 'virtualbox' provider...
==> default: Importing base box 'hashicorp/precise64'...
...
==> default: Forwarding ports...
    default: 22 (guest) => 2222 (host) (adapter 1)
==> default: Booting VM...
==> default: Waiting for machine to boot. This may take a few minutes...
    default: SSH address: 127.0.0.1:2222
...
==> default: Machine booted and ready!
```

```sh
$ vagrant provision
==> default: Running provisioner: ansible...
Windows is not officially supported for the Ansible Control Machine.
Please check https://docs.ansible.com/intro_installation.html#control-machine-requirements
    default: Running ansible-playbook...

PLAY ***************************************************************************

TASK [setup] *******************************************************************
ok: [default]

TASK [debug] *******************************************************************
ok: [default] => {
    "msg": "Hello Ansible from Vagrant"
}

PLAY RECAP *********************************************************************
default                    : ok=2    changed=0    unreachable=0    failed=0
```