Custom delpoyment tool for VMware ESXi Hypervisor
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
1. Download the latest release, and run ./go-via -f config.json
example config file
``` json
{
    "network": {
        "interfaces": ["ens224"]
    }
}
```
2. Download docker image from maxiepax/go-via:latest (not very tested!)

3. Download source and compile with go 1.15 and Angular 11

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

Known issue: when you press re-image, button does not change state to disabled.
Workaround: It does initiate the host to be re-imaged, if you open the developer tools and check console log, you will see that the database has been updated. Just reboot the host to be re-imaged and it will work.

Todo
-----

- [x] Fix progress bar when re-imaging hosts
- [] Fix re-image button so that it shows disabled once re-image has been initiated
- [] Fix log interface in UI so that logs can be viewed live
- [] Add on/off options to parts of default ks.cfg
- [] Add support for custom ks.cfg based on Group and Host
- [] Add more backend protection to not being able to remove Image/Groups/Pools that are in use by objects.
- [] Enhance default ks.cfg more, while still being secureboot compatible.
