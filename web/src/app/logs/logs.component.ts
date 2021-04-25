import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-logs',
  templateUrl: './logs.component.html',
  styleUrls: ['./logs.component.scss']
})
export class LogsComponent implements OnInit {

  data;

  constructor() {
    const ws = new Websocket('ws://' +  window.location.hostname + ':8080/v1/log')
    ws.addEventListener('message', event => {
      console.log(event.data);
    })
  }

  ngOnInit(): void {

  }
}


// 

