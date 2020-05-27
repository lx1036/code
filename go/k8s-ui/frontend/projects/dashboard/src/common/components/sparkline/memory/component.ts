

import {ChangeDetectionStrategy, Component, Input, OnInit} from '@angular/core';
import {MetricResult} from '@api/backendapi';
import {Sparkline} from '../sparkline';

@Component({
  selector: 'kd-memory-sparkline',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class MemorySparklineComponent extends Sparkline implements OnInit {
  @Input() timeseries: MetricResult[];

  ngOnInit() {
    this.setTimeseries(this.timeseries);
  }
}
