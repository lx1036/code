import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-pod-terminal',
  template: `

  `,
  styles: [
    `
          .terminal-parent {
              background: black;
              position: absolute;
              left: 0;
              top: 2.5rem;
              bottom: 0;
              right: 0;
              padding: 5px;
              width: 100%;
          }
    `
  ]
})
export class PodTerminalComponent implements OnInit {

  constructor() { }

  ngOnInit() {
  }

}
