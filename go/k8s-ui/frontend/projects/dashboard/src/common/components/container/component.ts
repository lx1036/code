

import {Component, Input, OnChanges} from '@angular/core';
import {ConfigMapKeyRef, Container, EnvVar, SecretKeyRef} from '@api/backendapi';
import {KdStateService} from '../../services/global/state';

@Component({
  selector: 'kd-container-card',
  templateUrl: './template.html',
  styleUrls: ['style.scss'],
})
export class ContainerCardComponent implements OnChanges {
  @Input() container: Container;
  @Input() namespace: string;
  @Input() initialized: boolean;

  constructor(private readonly state_: KdStateService) {}

  ngOnChanges(): void {
    this.container.env = this.container.env.sort((a, b) => a.name.localeCompare(b.name));
  }

  isSecret(envVar: EnvVar): boolean {
    return !!envVar.valueFrom && !!envVar.valueFrom.secretKeyRef;
  }

  isConfigMap(envVar: EnvVar): boolean {
    return !!envVar.valueFrom && !!envVar.valueFrom.configMapKeyRef;
  }

  formatSecretValue(s: string): string {
    return atob(s);
  }

  getEnvConfigMapHref(configMapKeyRef: ConfigMapKeyRef): string {
    return this.state_.href('configmap', configMapKeyRef.name, this.namespace);
  }

  getEnvSecretHref(secretKeyRef: SecretKeyRef): string {
    return this.state_.href('secret', secretKeyRef.name, this.namespace);
  }

  getEnvVarID(_: number, envVar: EnvVar): string {
    return `${envVar.name}-${envVar.value}`;
  }
}
