import { Component, OnInit } from '@angular/core';
import { ApiService } from '../api.service';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { cloudIcon, ClarityIcons } from '@cds/core/icon';
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


@Component({
  selector: 'app-manage-groups',
  templateUrl: './manage-groups.component.html',
  styleUrls: ['./manage-groups.component.scss']
})
export class ManageGroupsComponent implements OnInit {
  hosts;
  images;
  errors;
  groups;
  group;
  pools;
  Hostform: FormGroup;
  Groupform: FormGroup;
  groupid = null;
  poolid = null;
  showGroupModalMode = "";
  addHostFormModal = false;
  progress = {};
  progresstext = {};

  constructor(private apiService: ApiService, private HostformBuilder: FormBuilder, private GroupformBuilder: FormBuilder) {
    this.Hostform = this.HostformBuilder.group({
      fqdn: ['', [Validators.required]],
      ip: ['', [Validators.required]],
      mac: ['', [Validators.required]],
      group_id: ['', [Validators.required]],
    });
    this.Groupform = this.GroupformBuilder.group({
      name: ['', [Validators.required]],
      dns: ['', [Validators.required]],
      ntp: ['', [Validators.required]],
      password: ['', [Validators.required]],
      image_id: ['', [Validators.required]],
      pool_id: ['', [Validators.required]],
      erasedisks: [''],
      bootdisk: [''],
      ntp_pc: [''],
      ssh_pc: [''],
      domain_pc: [''],
    });
    const ws = new WebSocket('wss://' +  window.location.hostname + ':8443/v1/log')
    ws.addEventListener('message', event => {
      const data = JSON.parse(event.data)
      if (data.msg === "progress") {
        this.progress[data.id] = data.percentage;
        this.progresstext[data.id] = data.progresstext;
      }
    })
  }

  ngOnInit(): void {
    this.apiService.getGroups().subscribe((groups: any) => {
      this.apiService.getHosts().subscribe((hosts: any) => {
        this.groups = groups.map(item => {
          item.hosts = hosts.filter(host => host.group_id === item.id)
          return item
        });
        console.log(groups);
        hosts.forEach(host => {
          this.progress[host.id] = host.progress;
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

  submit_group() {
    const data = {
      ...this.Groupform.value,
      image_id: parseInt(this.Groupform.value.image_id),
      pool_id: parseInt(this.Groupform.value.pool_id),
    }

    let json_pc: any = {}
    if (data.ssh_pc) {
      json_pc.ssh = true;
      console.log('ssh is true');
    }
    if (data.ntp_pc) {
      json_pc.ntp = true;
    }
    if (data.domain_pc) {
      json_pc.domain = true;
    }
    if (data.erasedisks) {
      json_pc.erasedisks = true;
    }
    if (data.bootdisk) {
      json_pc.bootdisk = data.bootdisk;
    }

    data.options = json_pc;
    delete data.ssh_pc;
    delete data.ntp_pc;
    delete data.domain_pc;
    delete data.erasedisks;
    delete data.bootdisk;

    this.apiService.addGroup(data).subscribe((data: any) => {
      if (data.id) {
        this.groups.push(data);
        this.Groupform.reset();
        this.showGroupModalMode = '';
      }
    }, (data: any) => {
      if (data.status) {
        this.errors = data.error;
      } else {
        this.errors = [data.message];
      }

    });
    console.log(data);
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

  showGroupModal(mode, id=null) {
    this.showGroupModalMode = mode;
    if (mode === "edit") {
    this.group = this.groups.find(group => group.id === id);
    this.Groupform.patchValue({
      ...this.group,
      domain_pc: this.group.options.domain,
      ssh_pc: this.group.options.ssh,
      ntp_pc: this.group.options.ntp,
      erasedisks: this.group.options.erasedisks,
      bootdisk: this.group.options.bootdisk,
    });
    }
    if (mode === "add") {
      this.Groupform.reset();
    }
  }

  updateGroup() {
    const data = {
      ...this.Groupform.value,
      image_id: parseInt(this.Groupform.value.image_id),
      pool_id: parseInt(this.Groupform.value.pool_id),
    };

    let json_pc: any = {}
    if (data.ssh_pc) {
      json_pc.ssh = true;
      console.log('ssh is true');
    }
    if (data.ntp_pc) {
      json_pc.ntp = true;
    }
    if (data.domain_pc) {
      json_pc.domain = true;
    }
    if (data.erasedisks) {
      json_pc.erasedisks = true;
    }
    if (data.bootdisk) {
      json_pc.bootdisk = data.bootdisk;
    }

    data.options = json_pc;
    delete data.ssh_pc;
    delete data.ntp_pc;
    delete data.domain_pc;
    delete data.erasedisks;
    delete data.bootdisk;

    this.apiService.updateGroup(this.group.id, data).subscribe((resp: any) => {
      console.log(resp);
      if (resp.error) {
        this.errors = resp.error;
      }
      if (resp) {
        this.showGroupModalMode = '';
      }
    });
  }

  addHostToGroup(group_id, pool_id) {
    this.addHostFormModal = true;
    this.groupid = group_id;
    this.poolid = pool_id;
  }


  submitHost() {

    const data = {
      ...this.Hostform.value,
      hostname: this.Hostform.value.fqdn.split(".")[0],
      domain: this.Hostform.value.fqdn.split(".").slice(1).join('.'),
      group_id: parseInt(this.groupid),
      pool_id: parseInt(this.poolid),
    }

    this.apiService.addHost(data).subscribe((data: any) => {

      if (data.id) {
        const g = this.groups.find(group => group.id === data.group_id)
        g.hosts = [...(g.hosts || []) ,data]
        this.Hostform.reset();
        this.addHostFormModal = false;
      }
    }, (data: any) => {
      if (data.status) {
        this.errors = data.error;
      } else {
        this.errors = [data.message];
      }
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
    this.apiService.patchHost(id).subscribe((data: any) => {
      console.log("PUT Request is successful ", data);
      this.groups = this.groups.map(group => {
        group.hosts = group.hosts.map(host => host.id === id ? data : host)
        return group;
      })
      this.progress[id] = 0;
      this.progresstext[id] = "reimaging";
    },
    error  => {
      console.log("Error", error);
    });
  }
}