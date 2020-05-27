

import {EventEmitter, Injectable} from '@angular/core';
import {ObjectMeta, TypeMeta} from '@api/backendapi';

export class ResourceMeta {
  displayName: string;
  objectMeta: ObjectMeta;
  typeMeta: TypeMeta;

  constructor(displayName: string, objectMeta: ObjectMeta, typeMeta: TypeMeta) {
    this.displayName = displayName;
    this.objectMeta = objectMeta;
    this.typeMeta = typeMeta;
  }
}

@Injectable()
export class ActionbarService {
  onInit = new EventEmitter<ResourceMeta>();
  onDetailsLeave = new EventEmitter<void>();
}
