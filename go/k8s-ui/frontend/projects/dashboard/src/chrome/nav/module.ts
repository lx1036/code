

import {NgModule} from '@angular/core';
import {ComponentsModule} from '../../common/components/module';

import {NavServiceModule} from '../../common/services/nav/module';
import {SharedModule} from '../../shared.module';

import {NavComponent} from './component';
import {HamburgerComponent} from './hamburger/component';
import {NavItemComponent} from './item/component';
import {PinnerNavComponent} from './pinner/component';

@NgModule({
  declarations: [NavComponent, NavItemComponent, HamburgerComponent, PinnerNavComponent],
  exports: [NavComponent, NavItemComponent, HamburgerComponent],
  imports: [SharedModule, ComponentsModule, NavServiceModule],
})
export class NavModule {}
