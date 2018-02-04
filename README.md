# Ansible Wrapper for Windows

This is a simple executable wrapper to allow Vagrant to run ansible commands on windows through to a configured cygwin environment. There are a few [threads](http://stackoverflow.com/questions/29743491/how-to-install-ansible-playbook-on-windows-as-host-with-vagrant) out there that document the use of batch files to bootstrap ansible but I found the windows command shell does not actually work properly when using `--extra-vars` for instance.

We assume you have installed a working cygwin shell on your workstation.

## Wrapper Configuration

The wrapper must be able to find the installed cygwin environment. By default it assumes the cygwin home is `c:\cygwin`. If your cygwin is installed in a different location there are 2 options available to tell the wrapper where to find the cygwin environment.

1. Using the `CYGWIN_HOME` Environment Variable

```sh
export CYGWIN_HOME=c:\\cygwin64
```

2. Using the `cygwin.ini` File

The wrapper will look into the directory of the wrapper executable for a `cygwin.ini` file that contains the below ini information.

```ini
[cygwin]
home=c:\cygwin64
```
## Installing Ansible on Cygwin

This section describes the installation steps required to install Cygwin and Ansible on Windows.

### Install Cygwin

Use the cygwin package installer from https://cygwin.com/ and install the following packages:

```sh
setup-x86_64.exe -q -P binutils,^
                       gcc-core,^
                       make,^
                       python,^
                       python-setuptools,^
                       python-devel,^
                       libtool,^
                       libuuid-devel,^
                       libffi-devel,^
                       python-cffi,^
                       libcrypt-devel,^
                       python-crypto,^
                       openssl,^
                       openssl-devel,^
                       python-openssl,^
                       openssh,^
                       gmp,^
                       libgmp-devel,^
                       curl,^
                       wget,^
                       nano
```

or using `apt-cyg`

```sh
$ apt-cyg install binutils \
          gcc-core \
          make \
          python \
          python-setuptools \
          python-devel \
          libtool \
          libuuid-devel \
          libffi-devel \
          python-cffi \
          libcrypt-devel \
          python-crypto \
          openssl \
          openssl-devel \
          python-openssl \
          openssh \
          gmp \
          libgmp-devel \
          curl \
          wget \
          nano
```

Now you should have a basic cygwin python environment that can be used to install ansible in.

### Configure Python Environment

First lets open up a cygwin terminal and install `pip`

```sh
$ easy_install-2.7 pip
```

Thereafter we will update `pip`, `setuptools` and also install `wheel` and `virtualenv`.

```sh
$ pip install -U pip setuptools wheel virtualenv
```

### Install Ansible

Before installing Ansible, you will need to install its dependency `PyNaCL` with additional compiler flags.

```sh
ARCHFLAGS="-arch x86_64" \
LDFLAGS="-L/usr/lib" \
CFLAGS="-I/usr/include/openssl -g -O2 -D_BSD_SOURCE" pip install -v -v -v -U pynacl
```

Now lets install ansible and its dependencies.

```sh
$ pip install -U ansible
```

Quick test to ensure it works.

```sh
$ ansible --version
ansible 2.4.3.0
  ansible python module location = /usr/lib/python2.7/site-packages/ansible
  executable location = /usr/bin/ansible
  python version = 2.7.14 (default, Oct 31 2017, 21:12:13) [GCC 6.4.0]
```

### Configure Vagrant Environment to work with Ansible

To bridge Vagrant on windows to Ansible on Cygwin we can now use the Ansible Wrapper for Windows binary. The Vagrant Ansible provisionser will use the Vagrant windows environment to load the `ansible-playbook` executable. So we need to ensure Vagrant will find the wrapper executable to hand off to the Cygwin envionment.

Simply copy the `ansible-win-wrapper.exe` into the same folder of the `vagrant.exe` and rename it to `ansible-playbook.exe`.

```cmd
copy ansible-win-wrapper C:\HashiCorp\Vagrant\bin\ansible-playbook.exe

c:\> vagrant --version
Vagrant 2.0.1

c:\> ansible-playbook --version
ansible-playbook 2.4.3.0
  config file = None
  configured module search path = [u'/cygdrive/c/Users/fred/.ansible/plugins/modules', u'/usr/share/ansible/plugins/modules']
  ansible python module location = /usr/lib/python2.7/site-packages/ansible
  executable location = /usr/bin/ansible-playbook
  python version = 2.7.14 (default, Oct 31 2017, 21:12:13) [GCC 6.4.0]

```

Et voilà, ansible-playbook should now work from vagrant.

Make sure that when you install vagrant plugins to use the command line shell else the ruby gems from your Cygwin environment and from the vagrant embedded msys environment get all mixed up. Check that your Windows command shell `PATH` variable contains the `c:\HashiCorp\Vagrant\bin` path for vagrant to work correctly.

Below an example of how to install 2 of the most commonly used plugins.

```batch
c:\> vagrant plugin install vagrant-vbguest

c:\> vagrant plugin install vagrant-hostmanager
```

### Fix Vagrant SSH Bug

Some SSH arguments used by ansible playbook shell commands are hardcoded in Vagrant and cannot be overridden by `ansible.cfg` settings. In particular it will override the `control_path = none` setting in `ansible.cfg` with a hardcoded `-o ControlMaster=auto` shell ssh parameter.

As of 2018 setting up persistent ssh transactions is [not possible on Windows](http://stackoverflow.com/questions/20959792/is-ssh-controlmaster-with-cygwin-on-windows-actually-possible). To disable ControlMaster, either comment out line 314 in `$VAGRANT_HOME/embedded/gems/gems/vagrant-2.0.1/plugins/provisioners/ansible/provisioner/host.rb` or change it to `ControlMaster=none` or to `ControlMaster=no` for OpenSSH SSH v7.+ executables.

```rb
# which are lost when ANSIBLE_SSH_ARGS is defined.
unless ssh_options.empty?
  ssh_options << "-o ControlMaster=no"
  # DISABLE ssh_options << "-o ControlPersist=60s"
  # Intentionally keep ControlPath undefined to let ansible-playbook
  # automatically sets this option to Ansible default value
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
# default inventory file
inventory = ./inventory

# disable retry
retry_files_enabled = False

# disable host checking
host_key_ckecking = False
record_host_keys = False

[ssh_connection]
# below does not seem to work anymore so adding ssh args as is
ssh_args = -o ControlMaster=no -o UserKnownHostsFile=/dev/null -oStrictHostKeyChecking=no -o IdentitiesOnly=yes

# ControlMaster on cygwin OpenSSH does not work, must disable it
# otherwise ansible will not be able to connect to the target host
# control_path = none
```

**`inventory`**

Need to create an ansible inventory definition, the default Vagrant generated one does not work, vagrants ansible provisioner generates a inventory path arg that is Windows but must POSIX for ansible to work properly, so we just cheat a little.

```properties
default ansible_ssh_host=127.0.0.1 ansible_ssh_port=2222 ansible_ssh_user=vagrant ansible_ssh_private_key_file=./.vagrant/machines/default/virtualbox/private_key
```

Note: You may need to adjust the `ansible_ssh_port` to whatever the Vagrant machine bound to. You can get around all this by using a DNS plugin like `vagrant-hostmanager` and then just define actual DNS resolvable vagrant machine names in the inventory.

Also may need to `$ chmod 0644 inventory` depending on how your cygwin fstab is setup, ansible may complain about +x permissions.

**`playbook.yml`**

```yaml
---
- hosts: all
  tasks:
  - name: Greetings
    debug: msg="Hello Ansible from Vagrant {{ ansible_system }} {{ ansible_distribution }}"
  - name: Whats under the Hood
    command: uname -a
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

PLAY [all] *******************************************************************************************************************************************************************************************************

TASK [Gathering Facts] *******************************************************************************************************************************************************************************************
ok: [default]

TASK [Greetings] *************************************************************************************************************************************************************************************************
ok: [default] => {
    "msg": "Hello Ansible from Vagrant Linux Mandriva"
}

TASK [Whats under the Hood] **************************************************************************************************************************************************************************************
changed: [default] => {"changed": true, "cmd": ["uname", "-a"], "delta": "0:00:00.001658", "end": "2018-02-04 04:06:48.369022", "rc": 0, "start": "2018-02-04 04:06:48.367364", "stderr": "", "stderr_lines": [], "stdout": "Linux precise64 3.2.0-23-generic #36-Ubuntu SMP Tue Apr 10 20:39:51 UTC 2012 x86_64 x86_64 x86_64 GNU/Linux", "stdout_lines": ["Linux precise64 3.2.0-23-generic #36-Ubuntu SMP Tue Apr 10 20:39:51 UTC 2012 x86_64 x86_64 x86_64 GNU/Linux"]}

PLAY RECAP *******************************************************************************************************************************************************************************************************
default                    : ok=3    changed=1    unreachable=0    failed=0
```

### Random Historic Notes

#### Installing Crypto Dependencies

There also used to be a problem installing Ansible's dependency `pycrypto`. If you have an up-to-date version of Cygwin and its packages installed, you are fine. However if there are problems with `pycrypto`, here 2 options to install it on Cygwin.

Install pycrypto via pip using below compiler flag:

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

#### Vagrant Ansible Provisioner Bug Fix (version 1.8.x only)

This one is not to be missed, Vagrant does not send extra args in json format properly to ansible on Windows. Some discussion documented [here](https://github.com/mitchellh/vagrant/issues/6726).

Basically when providing any additional playbook parameters as a hash in Vagrant using the `extra_args` configuration option, this data is not sent in the right format to the ansible playbook.

```rb
ansible.extra_vars = {
  my_special_param: Array.new(2){ |n| "#{(1 + n).to_s.rjust(2,'0')}" }
}
```

Expected:

`--extra-vars='{ \"my_special_param\":[\"01\",\"02\"] }'`

Actual:

`--extra-vars={my_special_param:[01,02]}`

This problem is caused by the way the JSON is generated for a ruby hash.

The first problem is that the generated json does not escape the double quotes which will get lost in the Subprocess call out.

The other problem is that `ansible-playbook` is very picky in the way it receives and parses the json in the `--extra-vars` argument. In particular it expects a space after the opening bracket `{` and a space before the closig bracket `}`.


What did the trick for me was to apply this hack to the following vagrant file `$VAGRANT_HOME/embedded/gems/gems/vagrant-1.8.7/plugins/provisioners/ansible/provisioner/base.rb`, obviously this will void your warranty but it works.

```rb
def extra_vars_argument
  if has_an_extra_vars_file_argument
    # A JSON or YAML file is referenced.
    config.extra_vars
  else
    # Expected to be a Hash after config validation.
    config.extra_vars.to_json.gsub('{', '{ ').gsub('}', ' }').gsub('"', '\\\"') # << the hacked line fixes space of the curlies and escapes the double quotes
  end
end
```
