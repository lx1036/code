

import {Injectable} from '@angular/core';
import {PinnedResource} from '@api/backendapi';
import {HttpClient, HttpErrorResponse} from '@angular/common/http';
import {Subject} from 'rxjs';
import {MatDialog, MatDialogConfig} from '@angular/material/dialog';
import {AlertDialog, AlertDialogConfig} from '../../dialogs/alert/dialog';
import {VerberService} from './verber';

@Injectable()
export class PinnerService {
  onPinUpdate = new Subject();
  private isInitialized_ = false;
  private pinnedResources_: PinnedResource[] = [];
  private readonly endpoint_ = `api/v1/settings/pinner`;

  constructor(
    private readonly dialog_: MatDialog,
    private readonly http_: HttpClient,
    private readonly verber_: VerberService,
  ) {}

  init(): void {
    this.load();
    this.onPinUpdate.subscribe(() => this.load());
    this.verber_.onDelete.subscribe(() => this.load());
  }

  load(): void {
    this.http_.get<PinnedResource[]>(this.endpoint_).subscribe(resources => {
      this.pinnedResources_ = resources;
      this.isInitialized_ = true;
    });
  }

  isInitialized(): boolean {
    return this.isInitialized_;
  }

  pin(kind: string, name: string, namespace: string, displayName: string): void {
    this.http_
      .put(this.endpoint_, {kind, name, namespace, displayName})
      .subscribe(() => this.onPinUpdate.next(), this.handleErrorResponse_.bind(this));
  }

  unpin(kind: string, name: string, namespace: string): void {
    let url = `${this.endpoint_}/${kind}`;
    if (namespace !== undefined) {
      url += `/${namespace}`;
    }
    url += `/${name}`;

    this.http_
      .delete(url)
      .subscribe(() => this.onPinUpdate.next(), this.handleErrorResponse_.bind(this));
  }

  unpinResource(resource: PinnedResource): void {
    this.unpin(resource.kind, resource.name, resource.namespace);
  }

  isPinned(kind: string, name: string, namespace?: string): boolean {
    for (const pinnedResource of this.pinnedResources_) {
      if (
        pinnedResource.name === name &&
        pinnedResource.kind === kind &&
        pinnedResource.namespace === namespace
      ) {
        return true;
      }
    }
    return false;
  }

  getPinnedForKind(kind: string): PinnedResource[] {
    const resources = [];
    for (const pinnedResource of this.pinnedResources_) {
      if (pinnedResource.kind === kind) {
        resources.push(pinnedResource);
      }
    }

    return resources;
  }

  handleErrorResponse_(err: HttpErrorResponse): void {
    if (err) {
      const alertDialogConfig: MatDialogConfig<AlertDialogConfig> = {
        width: '630px',
        data: {
          title: err.statusText === 'OK' ? 'Internal server error' : err.statusText,
          message: err.error || 'Could not perform the operation.',
          confirmLabel: 'OK',
        },
      };
      this.dialog_.open(AlertDialog, alertDialogConfig);
    }
  }
}
