import {Component} from '@angular/core';
import {TranslateService} from '@ngx-translate/core';
import {CacheService} from "../../shared/components/auth/cache.service";
import {AuthService} from "../../shared/components/auth/auth.service";
import {StorageService} from "../../shared/common/storage.service";
import {SideNavCollapse} from "../../shared/common/sidenav-collapse";

@Component({
  selector: 'app-sidenav-namespace',
  templateUrl: './sidenav-namespace.component.html'
})
export class SidenavNamespaceComponent extends SideNavCollapse {
  collapsed = false;

  constructor(public authService: AuthService,
              public cacheService: CacheService,
              public translate: TranslateService,
              public storage: StorageService) {
    super(storage);
  }
}
