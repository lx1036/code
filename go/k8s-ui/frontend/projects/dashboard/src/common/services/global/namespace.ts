

import {EventEmitter, Injectable} from '@angular/core';
import {CONFIG} from '../../../index.config';

@Injectable()
export class NamespaceService {
  onNamespaceChangeEvent = new EventEmitter<string>();

  /**
   * Internal key for empty selection. To differentiate empty string from nulls.
   */
  private readonly allNamespacesKey_ = '_all';
  /**
   * Regular expression for namespace validation.
   */
  private readonly namespaceRegex = /^([a-z0-9]([-a-z0-9]*[a-z0-9])?|_all)$/;
  /**
   * Holds the currently selected namespace.
   */
  private currentNamespace_ = '';

  setCurrent(namespace: string) {
    this.currentNamespace_ = namespace;
  }

  current(): string {
    return this.currentNamespace_ || CONFIG.defaultNamespace;
  }

  getAllNamespacesKey(): string {
    return this.allNamespacesKey_;
  }

  getDefaultNamespace(): string {
    return CONFIG.defaultNamespace;
  }

  isNamespaceValid(namespace: string): boolean {
    return this.namespaceRegex.test(namespace);
  }

  isMultiNamespace(namespace: string): boolean {
    return namespace === this.allNamespacesKey_;
  }

  areMultipleNamespacesSelected(): boolean {
    return this.currentNamespace_ ? this.currentNamespace_ === this.allNamespacesKey_ : true;
  }
}
