import {Component, EmbeddedViewRef, NgModule, TemplateRef, ViewChild, ViewContainerRef} from '@angular/core';
import { CommonModule } from '@angular/common';
import {MatBottomSheet, MatBottomSheetConfig, MatBottomSheetModule, MatBottomSheetRef} from '../../packages/angular/material/bottom-sheet';
import {BrowserModule} from '@angular/platform-browser';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {Overlay, OverlayRef} from '@angular/cdk/overlay';
import {TemplatePortal} from '@angular/cdk/portal';


/**
 *
 *
 * https://blog.thoughtram.io/angular/2017/11/20/custom-overlays-with-angulars-cdk.html
 */


const defaultConfig = new MatBottomSheetConfig();


@Component({
  moduleId: module.id,
  selector: 'bottom-sheet-demo',
  styles: [
    `
          .demo-dialog-card {
              max-width: 405px;
              margin: 20px 0;
          }

          .mat-raised-button {
              margin-right: 5px;
          }
    `
  ],
  template: `
    <h1>Bottom sheet demo</h1>

    <button  (click)="openComponent()">Open component sheet</button>
    <button  (click)="openTemplate()">Open template sheet</button>

    
    <demo-overlay></demo-overlay>
    <!--<mat-card class="demo-dialog-card">-->
      <!--<mat-card-content>-->
        <!--<h2>Options</h2>-->

        <!--<p>-->
          <!--<mat-checkbox [(ngModel)]="config.hasBackdrop">Has backdrop</mat-checkbox>-->
        <!--</p>-->

        <!--<p>-->
          <!--<mat-checkbox [(ngModel)]="config.disableClose">Disable close</mat-checkbox>-->
        <!--</p>-->

        <!--<p>-->
          <!--<mat-form-field>-->
            <!--<input matInput [(ngModel)]="config.backdropClass" placeholder="Backdrop class">-->
          <!--</mat-form-field>-->
        <!--</p>-->

        <!--<p>-->
          <!--<mat-form-field>-->
            <!--<mat-select placeholder="Direction" [(ngModel)]="config.direction">-->
              <!--<mat-option value="ltr">LTR</mat-option>-->
              <!--<mat-option value="rtl">RTL</mat-option>-->
            <!--</mat-select>-->
          <!--</mat-form-field>-->
        <!--</p>-->

      <!--</mat-card-content>-->
    <!--</mat-card>-->


    <!--<ng-template let-bottomSheetRef="bottomSheetRef">-->
      <!--<mat-nav-list>-->
        <!--<mat-list-item (click)="bottomSheetRef.dismiss()" *ngFor="let action of [1, 2, 3]">-->
          <!--<mat-icon mat-list-icon>folder</mat-icon>-->
          <!--<span mat-line>Action {{ action }}</span>-->
          <!--<span mat-line>Description</span>-->
        <!--</mat-list-item>-->
      <!--</mat-nav-list>-->
    <!--</ng-template>-->
  `
})
export class BottomSheetDemo {
  config: MatBottomSheetConfig = {
    hasBackdrop: defaultConfig.hasBackdrop,
    disableClose: defaultConfig.disableClose,
    backdropClass: defaultConfig.backdropClass,
    direction: 'ltr'
  };

  @ViewChild(TemplateRef) template: TemplateRef<any>;

  constructor(private _bottomSheet: MatBottomSheet) {}

  openComponent() {
    this._bottomSheet.open(ExampleBottomSheet, this.config);
  }

  openTemplate() {
    this._bottomSheet.open(this.template, this.config);
  }
}

@Component({
  template: `
      <a href="#" (click)="handleClick($event)" *ngFor="let action of [1, 2, 3]">
        <span mat-line>Action {{ action }}</span>
        <span mat-line>Description</span>
      </a>
  `
})
export class ExampleBottomSheet {
  constructor(private sheet: MatBottomSheetRef) {}

  handleClick(event: MouseEvent) {
    event.preventDefault();
    this.sheet.dismiss();
  }
}





@Component({
  selector: 'demo-overlay',
  template: `
    
    <button (click)="open()">Open an overlay</button>
    <div #viewContainer></div>
    
    <ng-template #overlayTemplate let-person="test">
      <button>Open an Overlay Two</button>
      <div>
        <p>Hi {{person}}, Open an Overlay Two</p>
      </div>
    </ng-template>
  `
})
export class OverlayDemo {
  @ViewChild('overlayTemplate', {read: TemplateRef}) templateRef: TemplateRef<any>;
  @ViewChild('viewContainer', {read: ViewContainerRef}) viewContainer: ViewContainerRef;

  constructor(private _overlay: Overlay) {}

  open() {
    let overlayRef: OverlayRef = this._overlay.create({
      positionStrategy: this._overlay.position().global().bottom('100').centerVertically(),
      scrollStrategy: this._overlay.scrollStrategies.block(),
      hasBackdrop: true,
      width: '50%',
      height: '100%',
      direction: 'rtl',
    });

    console.log(this.templateRef);

    let templatePortal = new TemplatePortal(this.templateRef, this.viewContainer, {test: 'lx1036'});

    let embeddedViewRef: EmbeddedViewRef<{test: string}> = overlayRef.attach(templatePortal);

    console.log(embeddedViewRef.context, embeddedViewRef.rootNodes);
  }
}


@NgModule({
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    MatBottomSheetModule,
  ],
  declarations: [
    BottomSheetDemo,
    ExampleBottomSheet,
    OverlayDemo,
  ],
  entryComponents: [ // why ExampleBottomSheet register here, show detailed explanation???
    ExampleBottomSheet,
  ],
  bootstrap: [
    BottomSheetDemo
  ]
})
export class OverlayModule { }
