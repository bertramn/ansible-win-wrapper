# Ansible Wrapper for Windows

This is a simple executable wrapper to allow Vagrant to run ansible commands on windows through to a configured cygwin environment. There are a few [threads](http://stackoverflow.com/questions/29743491/how-to-install-ansible-playbook-on-windows-as-host-with-vagrant) out there that document the use of batch files to bootstrap ansible but I found the windows command shell does not actually work properly when using `--extra-vars` for instance.

We assume you have installed a working cygwin shell on your workstation.

### Installation

##### Install Cygwin Packages

Use the cygwin package installer from https://cygwin.com/ and install the following packages:
* binutils
* libuuid-devel
* python
* python-setuptools
* libffi-dev
* openssl
* openssl-devel
* libcrypt-devel
* gmp (optional)
* libgmp-devel (optional)
* gcc-core
* make
* openssh
* curl
* wget
* git
* nano (optional)

Now you should have a basic cygwin python environment that can be used to install ansible in.

##### Configure Python Environment

First lets open up a cygwin terminal and install `pip`

```sh
$ easy_install-2.7 pip
```

Thereafter we will update `pip` and `setuptools`

```sh
$ pip install -U pip setuptools
```
##### Install Ansible

Installing pycrypto is a bit of a drag as Ansible needs 2.6.1 and cygwin only comes with a precompiled 2.6 package. One can either install pycrypto via pip using below compiler flag:

```sh
CFLAGS="-g -O2 -D_BSD_SOURCE" pip install -U pycrypto
```

or just compile it from sources:

```sh
mkdir -p ~/workspaces/python && cd ~/workspaces/python
# download and unpack pycrypto 2.6.1 sources
curl https://ftp.dlitz.net/pub/dlitz/crypto/pycrypto/pycrypto-2.6.1.tar.gz | tar xzvf -

# compile disabling BSD source
cd pycrypto-2.6.1
CFLAGS="-g -O2 -D_BSD_SOURCE" python setup.py build build_ext -DMS_WIN64

# install module
python setup.py install
```

Still in the cygwin terminal lets install ansible and its dependencies.

```sh
$ pip install -U crypto paramiko PyYAML Jinja2 httplib2 six ansible
```

Quick test to ensure it works.

```sh
$ ansible --version
ansible 2.0.2.0
  config file =
  configured module search path = Default w/o overrides
```

### Configure Vagrant Environment to work with Ansible

To bridge Vagrant on windows to Ansible on Cygwin we can now use the Ansible Wrapper for Windows binary. The Vagrant Ansible provisionser will use the Vagrant windows environment to load the `ansible-playbook` executable. So we need to ensure Vagrant will find the wrapper executable to hand off to the Cygwin envionment.

Simply copy the `ansible-win-wrapper.exe` into the same folder of the `vagrant.exe` and rename it to `ansible-playbook.exe`.

```cmd
copy ansible-win-wrapper C:\HashiCorp\Vagrant\bin\ansible-playbook.exe

c:\> vagrant --version
Vagrant 1.8.1

c:\> ansible-playbook --version
ansible-playbook 2.0.1.0
  config file =
  configured module search path = Default w/o overrides
```

Et voilà, ansible-playbook should now work from vagrant.

Make sure that when you install vagrant plugins to use the command line shell else the ruby gems from your Cygwin environment and from the vagrant embedded msys environment get all mixed up. Check that your Windows command shell `PATH` variable contains the `c:\HashiCorp\Vagrant\bin` path for vagrant to work correctly.

Below an example of how to install 2 of the most commonly used plugins.

```cmd
c:\> vagrant plugin install vagrant-vbguest

c:\> vagrant plugin install vagrant-hostmanager
```

#### Some Advanced Hack

This one is not to be missed, Vagrant ( =< 1.8.1) does not send extra args in json format properly to ansible on Windows. Some discussion documented [here](https://github.com/mitchellh/vagrant/issues/6726).

What did the trick for me was to apply this hack to the following vagrant file `$VAGRANT_HOME/embedded/gems/gems/vagrant-1.8.1/plugins/provisioners/ansible/provisioner/base.rb`, obviously this will void your warranty but it works.

```rb
def extra_vars_argument
  if config.extra_vars.kind_of?(String) and config.extra_vars =~ /^@.+$/
    # A JSON or YAML file is referenced.
    config.extra_vars
  else
    # Expected to be a Hash after config validation.
    "'#{config.extra_vars.to_json.gsub("\"", %q(\\\"))}'" # <<<<< hack line
  end
end
```

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

**`ansible.cfg`**

Ansible default configuration with some Windows specific config.

```properties
[defaults]
host_key_ckecking = False
[ssh_connection]
# ControlMaster on cygwin OpenSSH does not work, must disable it
# otherwise ansible will not be able to connect to the target host
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
