import {
  AfterViewInit,
  Component,
  ElementRef,
  EventEmitter,
  Inject,
  OnDestroy,
  OnInit,
  Output,
  ViewChild
} from '@angular/core';
import {DOCUMENT} from '@angular/common';
import { animate, style, transition, trigger } from '@angular/animations';
import {CreateEditAppComponent} from "./create-edit-app/create-edit-app.component";
import {App} from "../../shared/models/app";
import {NamespaceClient} from "../../shared/controllers/kubernetes/namespace.service";
import {CacheService} from "../../shared/components/auth/cache.service";
import {MessageHandlerService} from "../../shared/message-handler.service";
import {AuthService} from "../../shared/components/auth/auth.service";

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

export const enum ActionType {
  ADD_NEW, EDIT
}




@Component({
  selector: 'app-portal-app',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],
  animations: [
    trigger('cardState', [
      transition('void => *', [
        style({opacity: 0, transform: 'translateY(-50%)'}),
        animate(200, style({opacity: 1, transform: 'translateY(0)'}))
      ]),
      transition('* => void', [
        animate(200, style({opacity: 0, transform: 'translateY(-50%)'}))
      ])
    ])
  ]
})
export class AppComponent implements OnInit, OnDestroy, AfterViewInit {
  showList: any[] = [];
  showState: object = showState;
  starredFilter: boolean;
  starredInherit: boolean; // starredInherit 用来传递给list
  resources: object = {};
  clusters: ClusterCard[] = [];
  allowNumber = 10;
  changedApps: App[];
  @ViewChild(CreateEditAppComponent, { static: false }) createEditApp: CreateEditAppComponent;

  constructor(private namespaceClient: NamespaceClient,
              private cacheService: CacheService,
              @Inject(DOCUMENT) private document: any,
              private element: ElementRef,
              private messageHandlerService: MessageHandlerService,
              public authService: AuthService) {}



  ngOnInit() {
    this.initShow();
    this.starredFilter = localStorage.getItem('starred') === 'true';
    this.starredInherit = this.starredFilter;
    // this.namespaceClient.getResourceUsage(this.cacheService.namespaceId).subscribe(response => {
    //   this.resources = response.data;
    //   Object.getOwnPropertyNames(this.resources).forEach(cluster => {
    //     this.clusters.push({name: cluster, state: false});
    //   });
    //
    //   this.allowNumber = this.getClusterMaxNumber();
    //   for (let i = 0; i < this.allowNumber - 1; i++) {
    //     setTimeout(((idx) => {
    //       if (this.clusters[idx]) {
    //         this.clusters[idx].state = true;
    //       }
    //     }).bind(this, i), 200 * i);
    //   }
    // }, error => this.messageHandlerService.handleError(error));
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

  openModal() {
    this.createEditApp.newOrEditApp();
  }

  createApp($event) {

  }

  editApp(app: App) {
    this.createEditApp.newOrEditApp(app.id);
  }
}
