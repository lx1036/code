

export class NamespaceMetaData {

}

export class Namespace {
  id: number;
  name: string;
  kubeNamespace: string;
  deleted: boolean;
  metaData: string;
  metaDataObj: NamespaceMetaData;
  user: string;
  createTime: Date;
  updateTime: Date;

  static emptyObject() {
    return undefined;
  }
}
