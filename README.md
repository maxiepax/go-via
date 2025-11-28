Custom deployment tool for VMware ESXi Hypervisor
=========================================

Credits
-------

Massive credits go to one of my best friends, and mentor [Jonathan "stamp" G](https://www.github.com/stamp) for all the help, coaching and lessons during this project.
Without your support this project would never have been a reality.

VMware #clarity-ui channel for being super helpful with newbie questions about clarity!


What is go-via?
---------------
go-via is a single binary, that when executed performs the tasks of dhcpd, tftpd, httpd, and ks.cfg generator, with a angular front-end, and http-rest backend written in go, and sqlite for persisting.

Why a new version of VMware Imaging Appliance?
----------------------------------------------
The old version of VIA had some things it didn't support which made it hard to run in enterprise environments. go-via brings added support for the following.
1. IP-Helper , you can have the go-via binary running on any network you want and use [RFC 3046 IP-Helper](https://tools.ietf.org/html/rfc3046) to relay DHCP requests to the server.
2. UEFI , go-via does not support BIOS, but does support UEFI and secure-boot. BIOS may be added in the future.
3. Virtual environments, it does not block nested esxi host deployment.
4. HTTP-REST, everything you can do in the UI, you can do via automation also.
5. Options to perform all prerequisites for VMware Cloud Foundation 4.x/5.x

Supported Architectures
-----------------------
UEFI x86_64 INTEL/AMD architecture
UEFI arm_64 ARM architecture (including Project Monterey/SmartNICs)

Default username / password / port
----------------------
username: admin <br>
password: VMware1!<br>
port: 8443<br>

Installation / Running
----------------------
<h3> Option 1: docker container </h3>
To run this container on a ubuntu 21.04 server, do the following:<br>

install docker-ce (https://docs.docker.com/engine/install/ubuntu/)
``` bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
```

install latest docker-compose,  
``` bash
sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/bin/docker-compose
sudo chmod +x /usr/bin/docker-compose
```

Option A: create the following docker-compose.yaml file to not specify a config file (dhcpd will serve on all interfaces)
``` yaml
version: "3.9"
services:
  go-via:
    image: tribock/go-via:latest
    network_mode: host
    volumes:
      - ./tftp:/go/tftp
      - ./database:/go/database
      - ./config:/go/config
      - ./cert:/go/cert
      - ./secret:/go/secret

```

Option B: or create this docker-compose.yaml to specify a config file, and place config in ./config/config.json
``` yaml
version: "3.9"
services:
  go-via:
    image: tribock/go-via:latest
    network_mode: host
    volumes:
      - ./tftp:/go/tftp
      - ./database:/go/database
      - ./config:/go/config
      - ./cert:/go/cert
      - ./secret:/go/secret
    command: -file /go/config/config.json

```
Example config file
``` json
{
    "network": {
        "interfaces": ["ens224", "ens192"]
    },
    "port": 443
}
```

now start the container

``` bash
sudo docker-compose up -d
```

<h3> Option 2: Download the latest release, and run ./go-via -file config.json </h3>

Most linux distributions should work, this has been tested on Ubuntu 20.20.

``` bash
#wget the release you want to download, e.g go-via_.<release>_linux_amd64.tar.gz
wget https://github.com/maxiepax/go-via/releases/download/<release>/go-via_.<release>_linux_amd64.tar.gz


#untar/extract it
tar -zxvf go-via_.<release>_linux_amd64.tar.gz
```
This will extract the files README.MD (this document) and go-via binary.

Optional: example config files.

Multi interface, and custom port.
``` json
{
    "network": {
        "interfaces": ["ens224", "ens192"]
    },
    "port": 443
}
```
Single interface, default port 8443
``` json
{
    "network": {
        "interfaces": ["ens224"]
    }
}
```

Now start the binary as super user, (optionally: pointing to the config file.)
``` bash
#start the application with default settings
sudo ./go-via

#start the application with normal debug level
sudo ./go-via -file config.json

#start the application with verbose debug level
sudo ./go-via -file config.json -debug
```

Example systemd.service config file
```
[Unit]
Description=go-via
After=network.target

[Service]
Type=simple
Restart=always
RestartSec=1
User=root
ExecStart=/home/govia/go-via
WorkingDirectory=/home/govia/go-via

[Install]
WantedBy=multi-user.target
```

You should be greeted with the following output.
``` bash
INFO[0000] Startup                                       commit=none date=unknown version=dev
WARN[0000] no interfaces have been configured, trying to find interfaces to serve to, will serve on all. 
INFO[0000] Existing database sqlite-database.db found   
INFO[0000] Starting dhcp server                          int=ens224 ip=172.16.100.1 mac="00:0c:29:91:cf:eb"
INFO[0000] Starting dhcp server                          int=ens192 ip=192.168.1.173 mac="00:0c:29:91:cf:e1"
INFO[0000] Starting dhcp server                          int=docker0 ip=172.17.0.1 mac="02:42:09:9f:04:4f"
INFO[0000] cert                                          server.crt="server.crt found"
INFO[0000] Webserver                                     port=":8443"
```

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

Troubleshooting
---------------
To troubleshoot, enable debugging.

Option 1: Docker Container. Append -debug to command.
``` bash
command: -debug
or
command: -file /go/config/config.json -debug
```

Option 2: Source or Binary. Append -debug to command
``` bash
./go-via -debug
```

Known issues
------------
Please note that go-via is still under heavy development, and there may be bugs. Following is the list of known issues.

currently tracking no known issues! :D


## Build

```bash
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

go install github.com/rakyll/statik@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/goreleaser/goreleaser@latest


docker buildx build --platform linux/amd64 --tag ghcr.io/tribock/go-via:latest --load .

docker run -p 8443:8443 ghcr.io/tribock/go-via:latest


go run -ldflags "-X main.commit=1.9 -X main.date=Heute"  *.go

# test auth
curl -X POST https://localhost:8443/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "VMware1!"}' -ivk

```


## Architecture

![Architecture](architecture/govia-overview.png)


Todo
-----
- Currently no requests have been made for features. Please submit any ideas you have.
