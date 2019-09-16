import {Component} from "@angular/core";


@Component({
  selector: 'dropdown-basic',
  template: `
    <ng-dropdown>
      <a ng-dropdown>Hover me</a>
      <ul ng-menu>
        <li ng-menu-item>
          <a>list1</a>
        </li>
        <li>
          <a>list2</a>
        </li>
        <li>
          <a>list3</a>
        </li>
      </ul>
    </ng-dropdown>
  `
})
export class DropdownDemo {

}
