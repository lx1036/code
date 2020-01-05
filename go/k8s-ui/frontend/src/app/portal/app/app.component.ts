import {AfterViewInit, Component, ElementRef, Inject, OnDestroy, OnInit} from '@angular/core';
import {NamespaceClient} from '../../shared/client/kubernetes/namespace';
import {CacheService} from '../../shared/auth/cache.service';
import {DOCUMENT} from '@angular/common';
import {MessageHandlerService} from '../../shared/message-handler/message-handler.service';

const showState = {
  name: {hidden: false},
  description: {hidden: false},
  create_time: {hidden: false},
  create_user: {hidden: false},
  action: {hidden: false}
};

interface ClusterCard {
  name: string;
  state: boolean;
}

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

  constructor(private namespaceClient: NamespaceClient,
              private cacheService: CacheService,
              @Inject(DOCUMENT) private document: any,
              private element: ElementRef,
              private messageHandlerService: MessageHandlerService) { }

  resources: object = {};
  clusters: ClusterCard[] = [];
  allowNumber = 10;

  ngOnInit() {
    this.initShow();
    this.starredFilter = localStorage.getItem('starred') === 'true';
    this.starredInherit = this.starredFilter;
    this.namespaceClient.getResourceUsage(this.cacheService.namespaceId).subscribe(response => {
      this.resources = response.data;
      Object.getOwnPropertyNames(this.resources).forEach(cluster => {
        this.clusters.push({name: cluster, state: false});
      });

      this.allowNumber = this.getClusterMaxNumber();
      for (let i = 0; i < this.allowNumber - 1; i++) {
        setTimeout(((idx) => {
          if (this.clusters[idx]) {
            this.clusters[idx].state = true;
          }
        }).bind(this, i), 200 * i);
      }
    }, error => this.messageHandlerService.handleError(error));
  }

  getClusterMaxNumber() {
    return Math.floor(this.element.nativeElement.querySelector('.cluster-outline').offsetWidth / 255);
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
