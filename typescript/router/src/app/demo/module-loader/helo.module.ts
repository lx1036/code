

import { NgModule, Component, OnInit, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'bitkan-hero',
  template: '<p>test</p>',
  // styleUrls: ['./hero.css']
})
export class BitkanHero implements OnInit {
  public name: string = 'hello word';
  public age: string = '40';
  @Output() onchange = new EventEmitter<any>();

  constructor() {}
  ngOnInit() {}
}

@NgModule({
  imports: [
    CommonModule
  ],
  declarations: [
    BitkanHero
  ],
  entryComponents: [
    BitkanHero
  ],
  providers: [
    { provide: 'ProvideBitkanHero', useClass: BitkanHero }
  ]
})
export class HeroModule {
}