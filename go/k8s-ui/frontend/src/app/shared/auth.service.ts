import {Injectable} from '@angular/core';
import {HttpClient, HttpErrorResponse} from '@angular/common/http';
import {CacheService} from './cache.service';
import {MessageHandlerService} from './message-handler.service';
import {TypePermission} from './model/v1/permission';
import {User} from './model/v1/user';

interface BaseConfig {
  appLabelKey: string;
  enableApiKeys: boolean;
  enableDBLogin: boolean;
  enableRobin: boolean;
  ldapLogin: boolean;
  oauth2Login: boolean;
}

@Injectable()
export class AuthService {
  config: BaseConfig;
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
    return this.http.get(`/me`).toPromise().then((response: {data: User}) => {
      this.currentUser = response.data;
      this.cacheService.setNamespaces(this.currentUser.namespaces);
      return response.data;
    });
  }

  initConfig(): Promise<BaseConfig> {
    return this.http
    .get(`/api/v1/configs/base`)
    .toPromise().then((response: {data: BaseConfig}) => {
        this.config = response.data;
        return response.data;
      }
    );
  }

  // Handle the related exceptions
  handleError(error: any): Promise<any> {
    return Promise.reject(error);
  }
}
