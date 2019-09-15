import {Component, NgModule} from '@angular/core';
import {NzAffixModule} from '../../packages/ng-zorro-antd/affix';
import {NzButtonModule} from '../../packages/ng-zorro-antd/button';
import {BrowserModule} from '@angular/platform-browser';








@Component({
  selector: `demo-affix`,
  template: `
    <nz-affix>
      <button nz-button [nzType]="'primary'">
        <span>Affix top</span>
      </button>
    </nz-affix>
    <br>
    <nz-affix nzOffsetBottom="0">
      <button nz-button [nzType]="'primary'">
        <span>Affix bottom</span>
      </button>
    </nz-affix>
  `
})
export class AffixComponent {
}



@NgModule({
  imports: [NzAffixModule, NzButtonModule, BrowserModule],
  exports: [],
  declarations: [AffixComponent],
  providers: [],
  bootstrap: [AffixComponent],
})
export class AffixModuleDemo {
}
