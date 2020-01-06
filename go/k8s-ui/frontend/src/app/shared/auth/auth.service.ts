import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {MessageHandlerService} from '../message-handler/message-handler.service';
import {CacheService} from './cache.service';
import {TypePermission} from '../model/v1/permission';


@Injectable()
export class AuthService {
  config: any;
  currentNamespacePermission: TypePermission = null;
  currentAppPermission: TypePermission = null;
  
  constructor(private http: HttpClient,
              private messageHandlerService: MessageHandlerService,
              public cacheService: CacheService) {
    this.currentAppPermission = new TypePermission();
    this.currentNamespacePermission = new TypePermission();
  }
  
  initConfig(): Promise<any> {
    return this.http
    .get(`/api/v1/configs/base`)
    .toPromise().then((response: any) => {
        this.config = response.data;
        return response.data;
      }
    ).catch(error =>
      this.handleError(error));
  }
  
  // Handle the related exceptions
  handleError(error: any): Promise<any> {
    // messageHandlerService
    return Promise.reject(error);
  }
}
