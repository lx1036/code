import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'kube-card',
  template: `
    <mat-card [ngClass]="{'kd-minimized-card': !expanded && !graphMode, 'kd-graph': graphMode}">
      <mat-card-title *ngIf="withTitle" (click)="expand()" [ngClass]="getTitleClasses()" fxLayoutAlign=" center">
        <div class="kd-card-title" fxFlex="100%">
          <ng-content select="[title]"></ng-content>
        </div>

        <div *ngIf="!expanded" class="kd-card-description kd-muted" fxLayoutAlign=" center" fxFlex="80">
          <ng-content select="[description]"></ng-content>
        </div>

        <div *ngIf="expanded && expandable" class="kd-card-actions kd-muted">
          <ng-content select="[actions]"></ng-content>
        </div>
        <mat-icon class="kd-card-toggle kd-muted" [matTooltip]="expanded ? 'Minimize card' : 'Expand card'" *ngIf="expandable">
          <ng-container *ngIf="expanded">arrow_drop_up</ng-container>
          <ng-container *ngIf="!expanded">arrow_drop_down</ng-container>
        </mat-icon>
      </mat-card-title>

      <div [@expandInOut]="!expanded">
        <mat-divider class="kd-on-top"></mat-divider>
        <mat-card-content class="kd-card-content" [ngClass]="{'kd-card-content-table': role === 'table'}">
          <div *ngIf="initialized; else loading">
            <ng-content select="[content]"></ng-content>
          </div>
          <ng-template #loading>
            <mat-progress-spinner mode="indeterminate"></mat-progress-spinner>
          </ng-template>
        </mat-card-content>
        <mat-card-footer>
          <div class="mat-small" [ngClass]="{'kd-card-footer': withFooter}">
            <ng-content select="[footer]"></ng-content>
          </div>
        </mat-card-footer>
      </div>
    </mat-card>
  `,
  styleUrls: ["./card.scss"]
})
export class CardComponent {
  @Input() initialized = true;
  @Input() role: string;
  @Input() withFooter = false;
  @Input() withTitle = true;
  @Input() expandable = true;
  @Input()
  set titleClasses(val: string) {
    this.classes_ = val.split(/\s+/);
  }
  @Input() expanded = true;
  @Input() graphMode = false;

  private classes_: string[] = [];

  expand(): void {
    if (this.expandable) {
      this.expanded = !this.expanded;
    }
  }


  getTitleClasses(): {[clsName: string]: boolean} {
    const ngCls = {} as {[clsName: string]: boolean};
    if (!this.expanded) {
      ngCls['kd-minimized-card-header'] = true;
    }

    if (this.expandable) {
      ngCls['kd-card-header'] = true;
    }

    for (const cls of this.classes_) {
      ngCls[cls] = true;
    }

    return ngCls;
  }
}

