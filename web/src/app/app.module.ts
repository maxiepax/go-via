import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { ClarityModule } from '@clr/angular';
import { CdsModule } from '@cds/angular';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ReactiveFormsModule } from '@angular/forms';
import { FormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { HelpComponent } from './help/help.component';
import { ManageDhcpPoolsComponent } from './manage-dhcp-pools/manage-dhcp-pools.component';
import { ManageGroupsComponent } from './manage-groups/manage-groups.component';
import { ManageImagesComponent } from './manage-images/manage-images.component';
import { LogsComponent } from './logs/logs.component';



@NgModule({
  declarations: [
    AppComponent,
    HelpComponent,
    ManageDhcpPoolsComponent,
    ManageGroupsComponent,
    ManageImagesComponent,
    LogsComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    ClarityModule,
    CdsModule,
    BrowserAnimationsModule,
    HttpClientModule,
    ReactiveFormsModule,
    FormsModule,
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
