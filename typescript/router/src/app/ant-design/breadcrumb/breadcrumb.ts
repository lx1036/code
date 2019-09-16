import {Component, Injector, Input, NgModule, NgZone, OnDestroy, OnInit, TemplateRef} from '@angular/core';
import {ActivatedRoute, PRIMARY_OUTLET, Router, RouterModule, Routes} from '@angular/router';
import {CommonModule} from '@angular/common';
import {BasicBreadcrumb} from './demo';
import {BrowserModule} from '@angular/platform-browser';
import {InputBoolean} from '../core/decorator';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {StringTemplateOutlet} from '../core/string_template_outlet';

@Component({
  selector: 'ng-breadcrumb',
  template: `
    <ng-content></ng-content>
    
    <ng-container *ngIf="autoGenerate">
      <ng-breadcrumb-item *ngFor="let breadcrumb of breadcrumbs">
        <a (click)="navigate(breadcrumb.url, $event)" [attr.href]="breadcrumb.url">{{breadcrumb.label}}</a>
      </ng-breadcrumb-item>
    </ng-container>
  `,
  styles: [
    `
      :host {
        display: block
      }
    `
  ]
})
export class BreadcrumbComponent implements OnInit, OnDestroy {
  public breadcrumbs = [];

  private destroy$ = new Subject<void>();

  @Input() @InputBoolean() autoGenerate = false;
  @Input() separator: string | TemplateRef<void> = '/';

  constructor(private _ngZone: NgZone, private _injector: Injector) {}

  public ngOnInit(): void {
    if (this.autoGenerate) {
      const activatedRoute = this._injector.get(ActivatedRoute);
      activatedRoute.url.pipe(takeUntil(this.destroy$)).subscribe(() => {
        this.breadcrumbs = this.getBreadCrumbs(activatedRoute.root);
      });
    }
  }

  public ngOnDestroy(): void {

  }

  public navigate(url: string, event: MouseEvent) {
    this._ngZone.run(() => {
      this._injector.get(Router).navigateByUrl(url).then();
    });
  }

  private getBreadCrumbs(root: ActivatedRoute, breadcrumbs: Array<any> = [], url: string = '') {
    const children: ActivatedRoute[] = root.children;

    if (children.length === 0) {
      return breadcrumbs;
    }

    for (const child of root.children) {
      if (child.outlet === PRIMARY_OUTLET) {
        const routeUrl = child.snapshot.url.map(segment => segment.path).join('/');
        const nextUrl = url + `/${routeUrl}`;

        if (routeUrl && child.snapshot.data.hasOwnProperty('breadcrumb')) {
          const breadcrumb = {
            url: nextUrl,
          };

          breadcrumbs.push(breadcrumb);
        }

        return this.getBreadCrumbs(child, breadcrumbs, nextUrl);
      }
    }
  }
}

@Component({
  selector: 'ng-breadcrumb-item',
  template: `
    <span class="breadcrumb-link">
      <ng-content></ng-content>
    </span>
    <span class="breadcrumb-separator">
      <ng-container *stringTemplateOutlet="breadcrumbComponent.separator">
        {{breadcrumbComponent.separator}}
      </ng-container>
    </span>
  `,
  styles: [
    `
      ng-breadcrumb-item:last-child {
      
      }
      
      ng-breadcrumb-item:last-child .breadcrumb-separator {
        display: none
      }
    `
  ]
})
export class BreadcrumbItemComponent {
  constructor(public breadcrumbComponent: BreadcrumbComponent) {}
}


const routes: Routes = [
  {
    path: '',
    component: BasicBreadcrumb
  }
];

@NgModule({
  imports: [BrowserModule, RouterModule.forRoot(routes)],
  declarations: [
    BreadcrumbComponent,
    BreadcrumbItemComponent,
    StringTemplateOutlet,

    BasicBreadcrumb,
  ],
  exports: [
    BreadcrumbComponent,
    BreadcrumbItemComponent,
  ],
  bootstrap: [
    BasicBreadcrumb,
  ]
})
export class BreadcrumbModule {}
