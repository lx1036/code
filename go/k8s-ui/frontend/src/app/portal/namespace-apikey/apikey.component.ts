import { Component, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { ClrDatagridStateInterface } from '@clr/angular';
import { Subscription } from 'rxjs/Subscription';
import {ListApiKeyComponent} from "./list-apikey/list-apikey.component";
import {CreateEditApiKeyComponent} from "./create-edit-apikey/create-edit-apikey.component";
import {ApiKey} from "../../shared/models/apikey";
import {PageState} from "../../shared/components/paginate/page-state";
import {ApiKeyService} from "./apikey.service";
import {CacheService} from "../../shared/components/auth/cache.service";
import {AuthService} from "../../shared/components/auth/auth.service";
import {MessageHandlerService} from "../../shared/components/notification/message-handler.service";
import {ConfirmationDialogService} from "../../shared/components/confirmation-dialog/confirmation-dialog.service";
import {ConfirmationState, ConfirmationTargets} from "../../shared/shared.const";

const showState = {
  'name': {hidden: false},
  'role': {hidden: false},
  'create_time': {hidden: false},
  'expire_time': {hidden: false},
  'create_user': {hidden: false},
  'description': {hidden: false},
  'action': {hidden: false}
};

@Component({
  selector: 'wayne-apikey.content-container',
  templateUrl: './apikey.component.html',
  styleUrls: ['./apikey.component.scss']
})
export class NamespaceApiKeyComponent implements OnInit, OnDestroy {
  @ViewChild(ListApiKeyComponent, { static: false })
  listApiKey: ListApiKeyComponent;
  @ViewChild(CreateEditApiKeyComponent, { static: false })
  createEditApiKey: CreateEditApiKeyComponent;
  changedApiKeys: ApiKey[];
  pageState: PageState = new PageState();
  showList: any[] = new Array();
  showState: object = showState;
  subscription: Subscription;

  constructor(private apiKeyService: ApiKeyService,
              public cacheService: CacheService,
              public authService: AuthService,
              private messageHandlerService: MessageHandlerService,
              private deletionDialogService: ConfirmationDialogService) {
    this.subscription = deletionDialogService.confirmationConfirm$.subscribe(message => {
      if (message &&
        message.state === ConfirmationState.CONFIRMED &&
        message.source === ConfirmationTargets.API_KEY) {
        const id = message.data;
        this.apiKeyService
          .deleteById(id, true, this.cacheService.namespaceId)
          .subscribe(
            response => {
              this.messageHandlerService.showSuccess('APIKey删除成功！');
              this.retrieve();
            },
            error => {
              this.messageHandlerService.handleError(error);
            }
          );
      }
    });
  }

  ngOnInit() {
    this.initShow();
  }

  initShow() {
    this.showList = [];
    Object.keys(this.showState).forEach(key => {
      if (!this.showState[key].hidden) { this.showList.push(key); }
    });
  }

  confirmEvent() {
    Object.keys(this.showState).forEach(key => {
      if (this.showList.indexOf(key) > -1) {
        this.showState[key] = {hidden: false};
      } else {
        this.showState[key] = {hidden: true};
      }
    });
  }

  cancelEvent() {
    this.initShow();
  }

  ngOnDestroy(): void {
    if (this.subscription) {
      this.subscription.unsubscribe();
    }
  }

  retrieve(state?: ClrDatagridStateInterface): void {
    if (state) {
      this.pageState = PageState.fromState(state, {
        totalPage: this.pageState.page.totalPage,
        totalCount: this.pageState.page.totalCount
      });
    }
    this.pageState.params['resourceId'] = this.cacheService.namespaceId;
    this.pageState.params['type'] = apiKeyTypeNamespace;
    this.pageState.params['deleted'] = false;
    this.pageState.sort.by = 'id';
    this.pageState.sort.reverse = true;
    this.apiKeyService.listPage(this.pageState, this.cacheService.namespaceId, null)
      .subscribe(
        response => {
          const data = response.data;
          this.pageState.page.totalPage = data.totalPage;
          this.pageState.page.totalCount = data.totalCount;
          this.changedApiKeys = data.list;
        },
        error => this.messageHandlerService.handleError(error)
      );
  }

  createApiKey(created: boolean) {
    if (created) {
      this.retrieve();
    }
  }

  openModal(): void {
    this.createEditApiKey.newOrEdit();
  }

  deleteApiKey(apiKey: ApiKey) {
    const deletionMessage = new ConfirmationMessage(
      '删除APIKey确认',
      '你确认删除APIKey ' + apiKey.name + ' ？',
      apiKey.id,
      ConfirmationTargets.API_KEY,
      ConfirmationButtons.DELETE_CANCEL
    );
    this.deletionDialogService.openComfirmDialog(deletionMessage);
  }

  editApiKey(apiKey: ApiKey) {
    this.createEditApiKey.newOrEdit(apiKey.id);
  }
}
