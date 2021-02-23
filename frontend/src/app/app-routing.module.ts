import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { ManageDhcpComponent } from './manage-dhcp/manage-dhcp.component';
import { ManageIsoComponent } from './manage-iso/manage-iso.component';
import { ManageHostsComponent } from './manage-hosts/manage-hosts.component';
import { HelpComponent } from './help/help.component';

const routes: Routes = [
 { path: 'manage-dhcp', component: ManageDhcpComponent },
 { path: 'manage-iso', component: ManageIsoComponent },
 { path: 'manage-hosts', component: ManageHostsComponent },
 { path: 'help', component: HelpComponent },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
