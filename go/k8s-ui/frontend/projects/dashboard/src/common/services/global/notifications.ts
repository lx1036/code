

import {Injectable} from '@angular/core';
import {K8sError} from '@api/backendapi';
import {GlobalSettingsService} from './globalsettings';

export class Notification {
  message: string;
  icon: string;
  cssClass: string;
  timestamp: Date;
  read = false;

  constructor(message: string, severity: NotificationSeverity) {
    this.message = message;
    this.icon = severity.toString();
    this.timestamp = new Date();

    switch (severity) {
      case NotificationSeverity.info:
        this.cssClass = 'kd-success';
        break;
      case NotificationSeverity.warning:
        this.cssClass = 'kd-warning';
        break;
      case NotificationSeverity.error:
        this.cssClass = 'kd-error';
        break;
      default:
        this.cssClass = '';
    }
  }
}

export enum NotificationSeverity {
  info = 'info',
  warning = 'warning',
  error = 'error',
}

@Injectable()
export class NotificationsService {
  private notifications_: Notification[] = [];

  constructor(private readonly _globalSettingsService: GlobalSettingsService) {}

  push(message: string, severity: NotificationSeverity): void {
    console.log(message);
    // Do not add same notifications multiple times
    if (this.notifications_.some(notification => notification.message === message)) {
      return;
    }

    this.notifications_ = [new Notification(message, severity), ...this.notifications_];
  }

  pushErrors(errors: K8sError[]): void {
    if (errors) {
      errors.forEach(error => {
        if (this._shouldAddNotification(error)) {
          this.push(error.ErrStatus.message, NotificationSeverity.error);
        }
      });
    }
  }

  private _shouldAddNotification(error: K8sError): boolean {
    return (
      !this._globalSettingsService.getDisableAccessDeniedNotifications() ||
      !this._isAccessDeniedError(error)
    );
  }

  private _isAccessDeniedError(error: K8sError): boolean {
    return error.ErrStatus.code === 403;
  }

  remove(index: number): void {
    this.notifications_.splice(index, 1);
  }

  getNotifications(): Notification[] {
    return this.notifications_;
  }

  getUnreadCount(): number {
    return this.notifications_
      .map(notification => {
        return notification.read ? Number(0) : Number(1);
      })
      .reduce((previousValue, currentValue) => {
        return previousValue + currentValue;
      }, 0);
  }

  markAllAsRead(): void {
    this.notifications_.forEach(notification => {
      notification.read = true;
    });
  }

  clear(): void {
    this.notifications_ = [];
  }
}
