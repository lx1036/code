import {Component} from "@angular/core";


@Component({
  selector: 'ng-dropdown',
  template: `
    <ng-content select="[ng-dropdown]"></ng-content>
    <ng-template>
      <div>
        <div>
          <ng-content select="[ng-menu]"></ng-content>
          <ng-content></ng-content>
        </div>
      </div>
    </ng-template>
  `
})
export class DropdownComponent {

}



