import { Component, OnInit } from '@angular/core';


@Component({
    selector: 'app-logs',
    templateUrl: './logs.component.html',
    styleUrls: ['./logs.component.scss'],
    standalone: false
})
export class LogsComponent implements OnInit {

  data = [];
  

  constructor() {
    const ws = new WebSocket('wss://' +  window.location.host + '/v1/log')
    ws.addEventListener('message', event => {
      const { time, level, msg, ...payload } = JSON.parse(event.data);
      const data = {
          time,
          level,
          msg,
          payload,
      };
      this.data.push(data);
    })
  }

  ngOnInit(): void {

  }
}


// 

