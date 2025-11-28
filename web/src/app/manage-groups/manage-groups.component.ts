import { Component, ElementRef, OnInit, ViewChild, ɵɵtrustConstantResourceUrl } from '@angular/core';
import { ApiService } from '../api.service';
import { FormGroup, UntypedFormBuilder, UntypedFormGroup, Validators } from '@angular/forms';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { ClarityIcons } from '@cds/core/icon';
import '@cds/core/icon/register.js';
import '@cds/core/accordion/register.js';
import '@cds/core/alert/register.js';
import '@cds/core/button/register.js';
import '@cds/core/checkbox/register.js';
import '@cds/core/datalist/register.js';
import '@cds/core/file/register.js';
import '@cds/core/forms/register.js';
import '@cds/core/input/register.js';
import '@cds/core/password/register.js';
import '@cds/core/radio/register.js';
import '@cds/core/range/register.js';
import '@cds/core/search/register.js';
import '@cds/core/select/register.js';
import '@cds/core/textarea/register.js';
import '@cds/core/time/register.js';
import '@cds/core/toggle/register.js';
import { id } from '@cds/core/internal';
import { GroupManagerComponent } from '../components/group-manager/group-manager.component';


@Component({
    selector: 'app-manage-groups',
    templateUrl: './manage-groups.component.html',
    styleUrls: ['./manage-groups.component.scss'],
    standalone: false
})
export class ManageGroupsComponent implements OnInit {
  hosts: any[] = [];
  host;
  groups: any[] = [];
  pools: any[] = [];
  images: any[] = [];
  errors: any;

  loadingFlag: boolean = false;

  group;

  advanced;
  Hostform: UntypedFormGroup;
  Groupform: UntypedFormGroup;
  groupid = null;
  poolid = null;
  showGroupModalMode = "";
  showHostModalMode = "";
  
  showVlanModalMode = "";
  vlan_id = null;

  HostVlanform: UntypedFormGroup;

  addHostFormModal = false;
  progress = {};
  progresstext = {};
  serverPowerStates = {}; // Add this property to track server power states

  visibleDatagrids: { [key: string]: boolean } = {};

  selectedGroup: any = {}; 

   @ViewChild(GroupManagerComponent) groupManager?: GroupManagerComponent;

  constructor(private apiService: ApiService, private HostformBuilder: UntypedFormBuilder, private GroupformBuilder: UntypedFormBuilder, private HostVlanformBuilder: UntypedFormBuilder) {
    
    this.Hostform = this.HostformBuilder.group({
      fqdn: ['', [Validators.required]],
      ip: ['', [Validators.required]],
      ilo_fqdn: [''],
      ilo_ip: [''],
      mac: ['', [Validators.required]],
      group_id: ['', [Validators.required]],
      ks: [''],
    });
    this.HostVlanform = this.HostVlanformBuilder.group({
      vlanID: ['', [Validators.required]],
      host_id: [''],
    });

    
    this.Groupform = this.GroupformBuilder.group({
      name: ['', [Validators.required]],
      dns: [''],
      ntp: [''],
      syslog: [''],
      vlan: [''],
      password: ['', [Validators.required]],
      image_id: ['', [Validators.required]],
      pool_id: ['', [Validators.required]],
      erasedisks: [''],
      bootdisk: [''],
      allowlegacycpu: [''],
      ssh: [''],
      certificate: [''],
      createvmfs: [''],
      callbackurl: [''],
      ks: [''],
    });
    const ws = new WebSocket('wss://' + window.location.host + '/v1/log')
    ws.addEventListener('message', event => {
      const data = JSON.parse(event.data)
      if (data.msg === "progress") {
        this.progress[data.id] = data.percentage;
        this.progresstext[data.id] = data.progresstext;
      }
      if (data.progresstext === "completed") {
        this.apiService.getGroups().subscribe((groups: any) => {
          this.apiService.getHosts().subscribe((hosts: any) => {
            this.groups = groups.map(item => {
              item.ks = atob(item.ks)
              item.hosts = hosts.filter(host => host.group_id === item.id) || []; // Ensure hosts array is initialized
              return item
            });
            hosts.forEach(host => {
              if (host.id === data.id) {
                host.reimage = false;
              }
            })
          });
        });
      }
    })
  }


  ngOnInit(): void {
    this.apiService.getGroups().subscribe((groups: any) => {
      this.apiService.getHosts().subscribe((hosts: any) => {
        console.log(groups, hosts)
        this.groups = groups.map(item => {
          item.ks = atob(item.ks)
          item.hosts = hosts.filter(host => host.group_id === item.id)
          return item
        });
        hosts.forEach(host => {
          this.progress[host.id] = host.progress;
          this.progresstext[host.id] = host.progresstext;
        })
      });
    });
    this.apiService.getImages().subscribe((images: any) => {
      this.images = images;
    });
    this.apiService.getPools().subscribe((pools: any) => {
      this.pools = pools;
    });
  }



  expandedGroupId: number | null = null;

  toggleDatagrid(groupId: number) {
    this.expandedGroupId = this.expandedGroupId === groupId ? null : groupId;

    
    // Only fetch power states when expanding a group
    if (this.expandedGroupId === groupId) {
      
      this.fetchPowerStatesForGroup(groupId);
    }
  }

  isDatagridVisible(groupId: number): boolean {
    return this.expandedGroupId === groupId;
  }

  fetchPowerStatesForGroup(groupId: number) {
    const group = this.groups.find(g => g.id === groupId);
    if (group && group.hosts) {
      // Clear existing power states for this group to force refresh
      group.hosts.forEach(host => {
        delete this.serverPowerStates[host.id];
      });
      
      // Fetch fresh power states for all hosts in the group
      group.hosts.forEach(host => {
        this.fetchPowerState(host.id);
      });
    }
  }

  refreshPowerStatesForGroup(groupId: number) {
    const group = this.groups.find(g => g.id === groupId);
    if (group && group.hosts) {
      // Clear existing power states for this group to force refresh
      group.hosts.forEach(host => {
        this.serverPowerStates[host.id] = 'refreshing...';
      });
      
      // Fetch fresh power states for all hosts in the group
      group.hosts.forEach(host => {
        this.fetchPowerState(host.id);
      });
    }
  }


  fetchPowerState(hostId: string) {
    var payload
    this.apiService.getHost(parseInt(hostId, 10)).subscribe((host: any) => {
      const id = host.id;
      var password = host.password;
      console.log(host)  
      payload = {
        iloIpAddr: host.ilo_ip,
        apiFlavour: host.ilo_api_flavour,
        port: host.ilo_port,
        username: host.ilo_user,
        password: host.ilo_password,
        vlanID: 123,
      };

      this.apiService.getServerPowerState(parseInt(hostId, 10), payload).subscribe(
        (response: any) => {
          // Success callback - this is where successful responses should be handled
          console.log("Server Power State for host", hostId, ":", response);
          this.serverPowerStates[hostId] = response.powerState || 'unknown';
        }, 
        (error: any) => {
          // Error callback - handle actual errors
          console.error("Error fetching server power state for host", hostId, ":", error);
          if (error.status) {
            this.errors = error.error;
          } else {
            this.errors = [error.message];
          }
          this.serverPowerStates[hostId] = 'error';
        }
      );

    });
  
  }
  

  removeGroup(id) {
    //check to see if group is empty
    var grp = this.groups.find(group => group.id === id);
    if (grp.hosts === undefined || grp.hosts.length == 0) {
      this.apiService.deleteGroup(id).subscribe((data: any) => {
        this.groups = this.groups.filter(group => group.id !== id);
      });
    } else {
      this.errors = ["The group is not empty, please delete all the hosts in the group first."];
    }
  }

  showGroupModal(mode, id = null) {
    this.showGroupModalMode = mode;
    if (mode === "edit") {
      this.group = this.groups.find(group => group.id === id);
      const { ssh, erasedisks, allowlegacycpu, certificate, createvmfs } = (this.group.options || {});
      this.Groupform.patchValue({
        ...this.group,
        ssh,
        erasedisks,
        allowlegacycpu,
        certificate,
        createvmfs,

      });
    }
    if (mode === "add") {
      this.Groupform.reset();
    }
  }

  showHostModal(mode, gid = null, hid = null) {
    this.showHostModalMode = mode;
    if (mode === "edit") {
      this.group = this.groups.find(group => group.id === gid);
      this.host = this.group.hosts.find(hosts => hosts.id === hid);
      var fqdn = this.host.hostname+"."+this.host.domain
      this.Hostform.patchValue({
        ...this.host,
        fqdn
      });
    }
    if (mode === "add") {
      this.Hostform.reset();
    }
  }





  updateGroup() {
    const data = {
      ...this.Groupform.value,
      image_id: parseInt(this.Groupform.value.image_id),
      pool_id: parseInt(this.Groupform.value.pool_id),
    };

    let json_pc: any = {}
    if (data.ssh) {
      json_pc.ssh = true;
    }
    if (data.erasedisks) {
      json_pc.erasedisks = true;
    }
    if (data.allowlegacycpu) {
      json_pc.allowlegacycpu = data.allowlegacycpu;
    }
    if (data.certificate) {
      json_pc.certificate = data.certificate;
    }
    if (data.createvmfs) {
      json_pc.createvmfs = data.createvmfs;
    }

    if (data.ks) {
      data.ks = btoa(data.ks)
    }

    // if no password has been entered, don't send it to avoid rehashing the hash.
    if (!data.password) {
      delete data.password;
    }

    data.options = json_pc;
    delete data.ssh;
    delete data.erasedisks;
    delete data.allowlegacycpu;
    delete data.certificate;
    delete data.createvmfs;

    this.apiService.updateGroup(this.group.id, data).subscribe((resp: any) => {
      delete resp.password;
      resp.ks = atob(resp.ks)
      this.groups = this.groups.map(group => {
        if (group.id === resp.id) {
          return { ...group, ...resp };
        }
        return group;
      });

      if (resp.error) {
        this.errors = resp.error;
      }
      if (resp) {
        this.errors = null;
        this.showGroupModalMode = '';
      }
    });
  }

  updateHost() {
    const data = {
      ...this.Hostform.value,
      hostname: this.Hostform.value.fqdn.split(".")[0],
      domain: this.Hostform.value.fqdn.split(".").slice(1).join('.'),
    };

    if (data.ks) {
      data.ks = btoa(data.ks)
    }

    this.apiService.updateHost(this.host.id, data).subscribe((resp: any) => {
      resp.ks = atob(resp.ks)

      this.groups = this.groups.map(group => {
        if (group.id === resp.group_id) {
          
          group.hosts = group.hosts.map(host => {
            if (host.id === resp.id) {
              host = resp;
            }
            return host;
          });

          return group;
        }
        return group;
      });
        
      if (resp.error) {
        this.errors = resp.error;
      }
      if (resp) {
        this.errors = null;
        this.showHostModalMode = '';
      }
    });
  }

  addHostToGroup(group_id, pool_id) {
    this.showHostModalMode = "add";
    this.groupid = group_id;
    this.poolid = pool_id;
  }

  @ViewChild('fileInput') fileInput!: ElementRef;

  triggerFileInput() {
    this.fileInput.nativeElement.click();
  }
    
  importHostsFromCSVToGroup(group_id, pool_id) {
      const target = event.target as HTMLInputElement;
    
      if (target.files && target.files.length > 0) {
        const file = target.files[0];
        console.log('Selected file:', file.name);
    
        // Call readCSVFile and handle the Promise
        this.readCSVFile(file).then((parsedData) => {
          console.log('Parsed Host Data:', parsedData);
          
          parsedData.forEach((host) => {
            const hostData = {
              fqdn: host.fqdn,
              ip: host.ip,   // Add other properties you need
              mac: host.mac, // If required
              hostname: host.fqdn.split(".")[0],
              domain: host.fqdn.split(".").slice(1).join('.'),
              group_id: parseInt(group_id),
              pool_id: parseInt(pool_id),
            };
    
            // Call addHost directly with the necessary data
            this.addHost(hostData); // Passing the data directly, not the form group           
          });
        }).catch((error) => {
          console.error('Error:', error);
          // Handle error, such as showing a notification to the user
        });
      }
    }
  readCSVFile(file: File): Promise<any[]> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();

      reader.onload = (e) => {
        const csvContent = e.target?.result as string;
        
        try {
          const parsedData = this.parseCSV(csvContent);
          resolve(parsedData); // Resolve with parsed data
        } catch (error) {
          reject('Error parsing CSV: ' + error); // Reject if there's an error parsing
        }
      };

      reader.onerror = (error) => {
        reject('Error reading file: ' + error); // Reject if there's an error reading the file
      };

      reader.readAsText(file);
    });
  }
    // Parses CSV string into an array of objects matching the form structure
    parseCSV(csv: string): any[] {
      const lines = csv.split('\n').map(line => line.trim()).filter(line => line);
      const headers = lines[0].split(',').map(h => h.trim()); // Extract headers
      
      const data: any[] = [];

      for (let i = 1; i < lines.length; i++) {
        const values = lines[i].split(',').map(v => v.trim());
        if (values.length === headers.length) {
          const formGroup = {};

          // Dynamically map values to the formGroup based on the header fields
          headers.forEach((header, index) => {
            let value = values[index] || ''; // Default to empty string if no value
            formGroup[header] = value;
          
          });

          // Push the formGroup into the data array
          data.push(formGroup);
        }
      }
      return data;
    }


  submitHost() {

    const data = {
      ...this.Hostform.value,
      hostname: this.Hostform.value.fqdn.split(".")[0],
      domain: this.Hostform.value.fqdn.split(".").slice(1).join('.'),
      group_id: parseInt(this.groupid),
      pool_id: parseInt(this.poolid),
    }
    this.addHost(data)

  }

  addHost(data){
    if (data.ks) {
      data.ks = btoa(data.ks)
    }
    this.apiService.addHost(data).subscribe((data: any) => {

      if (data.id) {
        const g = this.groups.find(group => group.id === data.group_id)
        g.hosts = [...(g.hosts || []), data]
        this.Hostform.reset();
        this.showHostModalMode = '';
        atob(data.ks)
      }
    }, (data: any) => {
      if (data.status) {
        this.errors = data.error;
      } else {
        this.errors = [data.message];
      }

      console.log(this.errors)
    });
  }

  removeHost(id) {
    this.apiService.deleteHost(id).subscribe((data: any) => {
      this.groups = this.groups.map(item => {
        item.hosts = item.hosts.filter(host => id !== host.id)
        return item;
      });
    });
  }

  reImageHost(id) {
    this.apiService.reimageHost(id).subscribe((data: any) => {
      this.groups = this.groups.map(group => {
        group.hosts = group.hosts.map(host => host.id === id ? data : host)
        return group;
      })
      this.progress[id] = 0;
      this.progresstext[id] = "reimaging";
    },
      error => {
        console.log("Error", error);
      });
  }

  cancelImageHost(id) {
    this.apiService.cancelImageHost(id).subscribe((data: any) => {
      this.groups = this.groups.map(group => {
        group.hosts = group.hosts.map(host => host.id === id ? data : host)
        return group;
      })
      this.progress[id] = 0;
      this.progresstext[id] = "reimaging canceled";
    },
      error => {
        console.log("Error", error);
      });
  }

  restartHost(hostId: string) {
    var payload
    this.apiService.getHost(parseInt(hostId, 10)).subscribe((host: any) => {
      const id = host.id;
      var password = host.password;
      console.log(host)  
      payload = {
        iloIpAddr: host.ilo_ip,
        apiFlavour: host.ilo_api_flavour,
        port: host.ilo_port,
        username: host.ilo_user,
        password: host.ilo_password,
        vlanID: 123,
      };

      this.apiService.restartHost(parseInt(hostId, 10), payload).subscribe((host: any) => {
  
      }, (response: any) => {
        if (response.status) {
          this.errors = response.error;
        } else {
          this.errors = [response.message];
        }
        this.loadingFlag = false; // Stop loading
      });

    

    });
  }

  shutdownHost(hostId: string) {
    var payload
    this.apiService.getHost(parseInt(hostId, 10)).subscribe((host: any) => {
      const id = host.id;
      var password = host.password;
      console.log(host)  
      payload = {
        iloIpAddr: host.ilo_ip,
        apiFlavour: host.ilo_api_flavour,
        port: host.ilo_port,
        username: host.ilo_user,
        password: host.ilo_password,
        vlanID: 123,
      };

      this.apiService.shutdownHost(parseInt(hostId, 10), payload).subscribe((host: any) => {
  
      }, (response: any) => {
        if (response.status) {
          this.errors = response.error;
        } else {
          this.errors = [response.message];
        }
        this.loadingFlag = false; // Stop loading
      });

    

    });
  }

  getServerPowerState(hostId: string) {
    // Use the centralized method instead of making direct API calls
    this.fetchPowerState(hostId);
    return this.serverPowerStates[hostId] || 'loading...';
  }

  startHost(hostId: string) {
    var payload
    this.apiService.getHost(parseInt(hostId, 10)).subscribe((host: any) => {
      const id = host.id;
      var password = host.password;
      console.log(host)  
      payload = {
        iloIpAddr: host.ilo_ip,
        apiFlavour: host.ilo_api_flavour,
        port: host.ilo_port,
        username: host.ilo_user,
        password: host.ilo_password,
        vlanID: 123,
      };

      this.apiService.startHost(parseInt(hostId, 10), payload).subscribe((host: any) => {
  
      }, (response: any) => {
        if (response.status) {
          this.errors = response.error;
        } else {
          this.errors = [response.message];
        }
        this.loadingFlag = false; // Stop loading
      });

    

    });

  }



  

  showVlanModal(hostId: string): void {
    console.log("HODOR!")
    console.log(hostId)
    this.HostVlanform.setValue({ 'host_id': hostId , 'vlanID': ''});
    this.showVlanModalMode = 'open'; // Open the modal
  }

  closeVlanModal(): void {
    this.showVlanModalMode = ''; // Close the modal
  }
  submitHostVlan(){
    this.loadingFlag = true; // Start loading
    this.apiService.getHost(parseInt(this.HostVlanform.get('host_id')?.value, 10)).subscribe((host: any) => {
      const id = host.id;
      var password = host.password;
      console.log(host)  
      const payload = {
        iloIpAddr: host.ilo_ip,
        apiFlavour: host.ilo_api_flavour,
        port: host.ilo_port,
        username: host.ilo_user,
        password: host.ilo_password,
        vlanID: this.HostVlanform.get('vlanID')?.value,
      };
      this.apiService.addHostVlan(id, payload).subscribe((response: any) => {
        this.groups = this.groups.map(group => {
          group.hosts = group.hosts.map(host => host.id === id ? response : host)
          return group;
        })
        this.showVlanModalMode = '';
        this.loadingFlag = false; // Stop loading
      }
      , (response: any) => {
        if (response.status) {
          this.errors = response.error;
        } else {
          this.errors = [response.message];
        }
        this.loadingFlag = false; // Stop loading
      }
      );
    });
  }

  onGroupSubmit(event: any) {
    if (this.showGroupModalMode === 'add') {
      this.groupManager.submitGroup().subscribe((data: any) => {
      if (data.id) {
        this.groups.push(data);
        this.Groupform.reset();
        this.showGroupModalMode = '';
        this.errors = null;
      }
    }, (data: any) => {
      if (data.status) {
        this.errors = data.error;
      } else {
        this.errors = [data.message];
      }

    });
  
    } else {
      this.updateGroup();
    }
  }
}