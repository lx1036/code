

import {Component, ElementRef, HostListener, OnInit} from '@angular/core';

import {Animations} from '../../common/animations/animations';
import {Notification, NotificationsService} from '../../common/services/global/notifications';

@Component({
  selector: 'kd-notifications',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
  animations: [Animations.easeOut],
})
export class NotificationsComponent {
  isOpen_ = false;
  notifications: Notification[] = [];

  constructor(
    private readonly notifications_: NotificationsService,
    private readonly element_: ElementRef,
  ) {}

  @HostListener('document:click', ['$event'])
  private onOutsideClick_(event: Event): void {
    if (!this.element_.nativeElement.contains(event.target) && this.isOpen()) {
      this.close_();
    }
  }

  load_(): void {
    this.notifications = this.notifications_.getNotifications();
  }

  open_(): void {
    this.load_();
    this.isOpen_ = true;
  }

  close_(): void {
    this.notifications_.markAllAsRead();
    this.isOpen_ = false;
  }

  isOpen(): boolean {
    return this.isOpen_;
  }

  toggle(): void {
    this.isOpen() ? this.close_() : this.open_();
  }

  remove(index: number): void {
    this.notifications_.remove(index);
  }

  clear(): void {
    this.notifications_.clear();
    this.load_();
  }

  getUnreadCount(): number {
    return this.notifications_.getUnreadCount();
  }
}
