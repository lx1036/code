

import {ObjectMeta, TypeMeta} from '@api/backendapi';

export class RawResource {
  static getUrl(typeMeta: TypeMeta, objectMeta: ObjectMeta): string {
    let resourceUrl = `api/v1/_raw/${typeMeta.kind}`;
    if (objectMeta.namespace !== undefined) {
      resourceUrl += `/namespace/${objectMeta.namespace}`;
    }
    resourceUrl += `/name/${objectMeta.name}`;
    return resourceUrl;
  }
}
