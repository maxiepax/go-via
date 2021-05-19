Custom deployment tool for VMware ESXi Hypervisor
=========================================

Credits
-------

Massive credits go to one of my best friends, and mentor [Jonathan "Stamp" Grimmtjärn](https://www.github.com/stamp) for all the help, coaching and lessons during this project.
Without your support this project would never been a reality.

VMware #clarity-ui channel for being super helpful with newbie questions about clarity!


What is go-via?
---------------
go-via is a single binary, that when executed performs the tasks of dhcpd, tftpd, httpd, and ks.cfg generator, with a angular front-end, and http-rest backend written in go, and sqlite for persisting.

Installation / Running
----------------------
<h3> Option 1: Download the latest release, and run ./go-via -file config.json </h3>

Most linux distributions should work, this has been tested on Ubuntu 20.20.

``` bash
#wget the release you want to download, e.g go-via_.0.0.25_linux_amd64.tar.gz
wget https://github.com/maxiepax/go-via/releases/download/v.0.0.25/go-via_.0.0.25_linux_amd64.tar.gz


#untar/extract it
tar -zxvf go-via_.0.0.24_linux_amd64.tar.gz
```
This will extract the files README.MD (this document) and go-via binary.

create an example config file, e.g. config.json, replace ens224 with the network interface you want to use to serve dhcpd/tftp.
``` json
{
    "network": {
        "interfaces": ["ens224"]
    }
}
```
Now start the binary as super user, pointing to the config file.
``` bash
#start the application with normal debug level
sudo ./go-via -file config.json

#start the application with verbose debug level
sudo ./go-via -file config.json -debug
```
You should be greeted with the following output.
``` bash
sudo ./go-via -file config.json 
INFO[0000] Startup                                       commit=bfe02f13d3382f1c760a1510fd3bbb966b5ac3f6 date="2021-04-26T12:01:33Z" version=.0.0.24
INFO[0000] Existing database sqlite-database.db found   
INFO[0000] Starting dhcp server                          int=ens224 ip=172.16.100.1 mac="00:0c:29:91:cf:eb"
INFO[0000] Webserver                                     address=":8080"
```
You can now browse to the web-frontend on the ip of the interface you specified, and the port 8080.

<h3> Option 2: docker container </h3>
todo: i automatically build a container each build, but havnt tested that it actually works, see this as a placeholder for now.

<h3> Option 3: Download source and compile with go 1.16 and Angular 11 </h3>

with Ubuntu 20.20 installed, do the following:
install golang 1.16.x compiler
``` bash
sudo snap install go --classic
```
install npm
``` bash
sudo apt-get install npm
```
install angular-cli
``` bash
sudo npm install npm@latest -g
sudo npm install -g @angular/cli
```
start two terminals:

terminal 1:
``` bash
mkdir ~/go
cd ~/go
git clone https://github.com/maxiepax/go-via.git
cd go-via
go run *.go
```

terminal 2:
``` bash
cd ~/go-via/web
npm install
# to only allow localhost access to gui:
ng serve
# to allow anyone access to gui:
ng serve --host 0.0.0.0
```

Why a new version of VMware Imaging Appliance?
----------------------------------------------
The old version of VIA had some things it didn't support which made it hard to run in enterprise environments. go-via brings added support for the following.
1. IP-Helper , you can have the go-via binary running on any network you want and use [RFC 3046 IP-Helper](https://tools.ietf.org/html/rfc3046) to relay DHCP requests to the server.
2. UEFI , go-via does not support BIOS, but does support UEFI and secure-boot. BIOS may be added in the future.
3. Custom ks.cfg files, you can use the embedded default or override on Group or Host level.
4. Virtual environments, it does not block nested esxi host deployment.
5. HTTP-REST, everything you can do in the UI, you can do via automation also.

Known issues
------------
Please note that go-via is still under heavy development, and there are bugs. Following is the list of known issues.

Known issue: When booting a host, it will request mboot.efi and successfully load it, however says it fails to load boot.cfg. Logs will show that it actually never requested boot.cfg.
Workaround: Just reboot the host, eventually it magically starts working.

Todo
-----
- [ ] Authentication (basicAuth)
- [ ] post-config: regenerate self-signed certificate with correct hostname
- [x] Fix progress bar when re-imaging hosts
- [x] Fix re-image button so that it shows disabled once re-image has been initiated
- [x] Fix log interface in UI so that logs can be viewed live
- [x] Add post-deployment configuration
- [x] Add support for custom ks.cfg based on Group and Host
- [x] Add more backend protection to not being able to remove Image/Groups/Pools that are in use by objects.
- [x] Enhance default ks.cfg more, while still being secureboot compatible. - note: ks.cfg is not really possible with secureboot, added option to do post-config via SOAP API instead.
