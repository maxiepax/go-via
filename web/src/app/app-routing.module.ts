import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { HelpComponent } from './help/help.component';
import { ManageDhcpPoolsComponent } from './manage-dhcp-pools/manage-dhcp-pools.component';
import { ManageGroupsComponent } from './manage-groups/manage-groups.component';
import { ManageImagesComponent } from './manage-images/manage-images.component';
import { LogsComponent } from './logs/logs.component';


const routes: Routes = [
  { path: 'manage-dhcp-pools', component: ManageDhcpPoolsComponent },
  { path: 'manage-groups', component: ManageGroupsComponent },
  { path: 'manage-images', component: ManageImagesComponent },
  { path: 'help', component: HelpComponent },
  { path: 'logs', component: LogsComponent },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
