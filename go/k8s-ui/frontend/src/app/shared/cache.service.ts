import {Injectable} from '@angular/core';
import {Namespace} from './model/v1/namespace';
import {MessageHandlerService} from './message-handler.service';

@Injectable()
export class CacheService {
  namespace: Namespace;
  namespaces: Namespace[];

  constructor(private messageHandlerService: MessageHandlerService) {
  }

  get namespaceId(): number {
    if (this.namespace) {
      return this.namespace.id;
    } else {
      this.messageHandlerService.error('当前用户无任何命名空间权限，请联系管理员添加！');
    }
  }

  setNamespaces(namespaces: Namespace[]) {
    this.namespaces = namespaces;
  }
}
