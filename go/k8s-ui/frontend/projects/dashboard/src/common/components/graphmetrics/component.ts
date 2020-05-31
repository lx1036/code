

import {Component, Input} from '@angular/core';
import {Metric} from '@api/backendapi';
import {GraphType} from '../graph/component';

@Component({
  selector: 'kd-graph-metrics',
  templateUrl: './template.html',
})
export class GraphMetricsComponent {
  @Input() metrics: Metric[];

  readonly GraphType: typeof GraphType = GraphType;

  showGraphs(): boolean {
    return (
      this.metrics &&
      this.metrics.every(metrics => metrics.dataPoints && metrics.dataPoints.length > 1)
    );
  }
}
