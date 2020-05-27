

import {ChangeDetectionStrategy, Component, Input, OnInit} from '@angular/core';
import {MetricResult} from '@api/backendapi';
import {Sparkline} from '../sparkline';

@Component({
  selector: 'kd-cpu-sparkline',
  templateUrl: './template.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class CpuSparklineComponent extends Sparkline implements OnInit {
  @Input() timeseries: MetricResult[];

  ngOnInit() {
    this.setTimeseries(this.timeseries);
  }
}
