

import {Component, Input, OnChanges, SimpleChanges} from '@angular/core';
import {Metric} from '@api/backendapi';
import {GraphType} from '../graph/component';

@Component({selector: 'kd-graph-card', templateUrl: './template.html'})
export class GraphCardComponent implements OnChanges {
  @Input() graphTitle: string;
  @Input() graphInfo: string;
  @Input() graphType: GraphType;
  @Input() metrics: Metric[];
  @Input() selectedMetricName: string;
  selectedMetric: Metric;

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['metrics']) {
      this.metrics = changes['metrics'].currentValue;
      this.selectedMetric = this.getSelectedMetrics();
    }
  }

  private getSelectedMetrics(): Metric {
    if (
      !this.selectedMetricName ||
      (this.metrics.length && this.metrics[0].dataPoints.length === 0)
    ) {
      return null;
    }

    return (
      this.metrics &&
      this.metrics.filter(metric => metric.metricName === this.selectedMetricName)[0]
    );
  }

  shouldShowGraph(): boolean {
    return this.selectedMetric !== undefined;
  }
}
