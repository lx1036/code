import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {CacheService} from './cache.service';
import {MessageHandlerService} from "./message-handler.service";
import {TypePermission} from "./model/v1/permission";
import {User} from "./model/v1/user";


@Injectable()
export class AuthService {
  config: any;
  currentNamespacePermission: TypePermission = null;
  currentAppPermission: TypePermission = null;
  currentUser: User = null;
  
  constructor(private http: HttpClient,
              private messageHandlerService: MessageHandlerService,
              public cacheService: CacheService) {
    this.currentAppPermission = new TypePermission();
    this.currentNamespacePermission = new TypePermission();
  }
  
  retrieveUser(): Promise<User> {
    return this.http.get(`/currentuser`).toPromise().then((response: any) => {
      this.currentUser = response.data;
      this.cacheService.setNamespaces(this.currentUser.namespaces);
      return response.data;
    }).catch((error) => {
      this.messageHandlerService.handleError(error);
      return Promise.resolve();
    });
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
    return Promise.reject(error);
  }
}
