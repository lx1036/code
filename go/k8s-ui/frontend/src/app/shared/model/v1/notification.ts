import {User} from './user';


export class Notification {
  id: number;
  type: string;
  title: string;
  message: string;
  from: User;
  level: number;
  isPublished: boolean;
  createTime: Date;
  updateTime: Date;

  constructor() {
    this.type = '公告';
    this.level = 0;
    this.title = '';
  }
}

export class NotificationLog {
  id: number;
  isRead: boolean;
  notification: Notification;
}

