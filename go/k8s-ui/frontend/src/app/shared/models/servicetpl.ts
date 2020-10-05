import {PublishStatus} from "./publish-status";

export class Servicetpl {
  id: number;
  name: string;
  serviceId: number;
  template: string;
  description: string;
  deleted: boolean;
  user: string;
  createTime: Date;
  updateTime?: Date;
  service: Service;

  ports: string;
  status: PublishStatus[];
}
