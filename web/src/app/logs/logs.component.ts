import { Component, OnInit } from '@angular/core';
import ReconnectingWebSocket from 'reconnecting-websocket';


@Component({
  selector: 'app-logs',
  templateUrl: './logs.component.html',
  styleUrls: ['./logs.component.scss']
})
export class LogsComponent implements OnInit {

  constructor() { }

  ngOnInit(): void {
    const rws = new ReconnectingWebSocket('ws://my.site.com');

    rws.addEventListener('open', () => {
    });

    rws.onmessage = function(event) {
      console.debug("WebSocket message received:", event);
    };
  }

}
