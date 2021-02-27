import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { ClarityModule } from '@clr/angular';
import { CdsModule } from '@cds/angular';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ReactiveFormsModule } from '@angular/forms';
import { HelpComponent } from './help/help.component';
import { ManageDhcpPoolsComponent } from './manage-dhcp-pools/manage-dhcp-pools.component';
import { ManageGroupsComponent } from './manage-groups/manage-groups.component';
import { ManageImagesComponent } from './manage-images/manage-images.component';


@NgModule({
  declarations: [
    AppComponent,
    HelpComponent,
    ManageDhcpPoolsComponent,
    ManageGroupsComponent,
    ManageImagesComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    ClarityModule,
    CdsModule,
    BrowserAnimationsModule,
    ReactiveFormsModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
