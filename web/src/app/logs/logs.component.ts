import { Component, OnInit } from '@angular/core';

import { webSocket } from "rxjs/webSocket";
const subject = webSocket('ws://' +  window.location.hostname + ':8080/v1/log');

subject.subscribe(
   msg => console.log('message received: ' + msg), // Called whenever there is a message from the server.
   err => console.log(err), // Called if at any point WebSocket API signals some kind of error.
   () => console.log('complete') // Called when connection is closed (for whatever reason).
 );

@Component({
  selector: 'app-logs',
  templateUrl: './logs.component.html',
  styleUrls: ['./logs.component.scss']
})
export class LogsComponent implements OnInit {

  constructor() { }

  ngOnInit(): void {


  }
}


