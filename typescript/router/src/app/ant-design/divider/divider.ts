import {ChangeDetectionStrategy, Component, OnChanges, OnInit, SimpleChanges} from "@angular/core";


@Component({
  selector: 'ng-divider',
  template: `
    <span *ngIf="ngText" class="ng-divider-inner-text">
      <ng-container *stringTemplateOutlet="ngText">{{ngText}}</ng-container>
    </span>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class DividerComponent implements OnInit, OnChanges {
  ngOnInit(): void {
    
  
  }
  
  ngOnChanges(changes: SimpleChanges): void {
  }
  
  

}
