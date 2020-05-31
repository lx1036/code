

import {HttpParams} from '@angular/common/http';
import {ChangeDetectionStrategy, ChangeDetectorRef, Component, Input, OnInit} from '@angular/core';
import {Event, EventList} from '@api/backendapi';
import {Observable} from 'rxjs/Observable';

import {ResourceListWithStatuses} from '../../../resources/list';
import {NotificationsService} from '../../../services/global/notifications';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {ListGroupIdentifier, ListIdentifier} from '../groupids';

const EVENT_TYPE_WARNING = 'Warning';

@Component({
  selector: 'kd-event-list',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class EventListComponent extends ResourceListWithStatuses<EventList, Event>
  implements OnInit {
  @Input() endpoint: string;

  constructor(
    private readonly eventList: NamespacedResourceService<EventList>,
    notifications: NotificationsService,
    cdr: ChangeDetectorRef,
  ) {
    super('', notifications, cdr);
    this.id = ListIdentifier.event;
    this.groupId = ListGroupIdentifier.none;

    // Register status icon handler
    this.registerBinding(this.icon.warning, 'kd-warning', this.isWarning);
    this.registerBinding(this.icon.none, '', this.isNormal.bind(this));
  }

  ngOnInit(): void {
    if (this.endpoint === undefined) {
      throw Error('Endpoint is a required parameter of event list.');
    }

    super.ngOnInit();
  }

  isWarning(event: Event): boolean {
    return event.type === EVENT_TYPE_WARNING;
  }

  isNormal(event: Event): boolean {
    return !this.isWarning(event);
  }

  getResourceObservable(params?: HttpParams): Observable<EventList> {
    return this.eventList.get(this.endpoint, undefined, undefined, params);
  }

  map(eventList: EventList): Event[] {
    return eventList.events;
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'message', 'source', 'subobject', 'count', 'firstseen', 'lastseen'];
  }
}
