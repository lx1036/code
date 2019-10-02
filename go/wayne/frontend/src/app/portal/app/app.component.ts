import {AfterViewInit, Component, OnDestroy, OnInit} from '@angular/core';
import {NamespaceClient} from "../../shared/client/kubernetes/namespace";

const showState = {
  'name': {hidden: false},
  'description': {hidden: false},
  'create_time': {hidden: false},
  'create_user': {hidden: false},
  'action': {hidden: false}
};

@Component({
  selector: 'wayne-app',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],
  animations: [
  
  ]
})
export class AppComponent implements OnInit, OnDestroy, AfterViewInit {
  showList: any[] = [];
  showState: object = showState;
  starredFilter: boolean;
  starredInherit: boolean; // starredInherit 用来传递给list
  
  constructor(private namespaceClient: NamespaceClient, private cacheService: CacheService) { }

  ngOnInit() {
    this.initShow();
    this.starredFilter = (localStorage.getItem('starred') === 'true');
    this.starredInherit = this.starredFilter;
    this.namespaceClient.getResourceUsage(this.cacheService.namespaceId).subscribe(response => {}, error => {});
  }
  
  initShow() {
    this.showList = [];
    Object.keys(this.showState).forEach(key => {
      if (!this.showState[key].hidden) {
        this.showList.push(key);
      }
    });
  }
  
  ngAfterViewInit(): void {
  }
  
  ngOnDestroy(): void {
  }
}
