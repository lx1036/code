

import {MetricResult} from '@api/backendapi';

export class Sparkline {
  lastValue = 0;

  private _timeseries: MetricResult[];

  setTimeseries(timeseries: MetricResult[]) {
    this._timeseries = timeseries;
  }

  getPolygonPoints(): string {
    const series = this._timeseries.map(({timestamp, value}) => [Date.parse(timestamp), value]);
    const sorted = series.slice().sort((a, b) => a[0] - b[0]);
    this.lastValue = sorted.length > 0 ? sorted[sorted.length - 1][1] : undefined;
    const xShift = Math.min(...sorted.map(pt => pt[0]));
    const shifted = sorted.map(([x, y]) => [x - xShift, y]);
    const xScale = Math.max(...shifted.map(pt => pt[0])) || 1;
    const yScale = Math.max(...shifted.map(pt => pt[1])) || 1;
    const scaled = shifted.map(([x, y]) => [x / xScale, y / yScale]);

    // Invert Y because SVG Y=0 is at the top, and we want low values
    // of Y to be closer to the bottom of the graphic.
    const map = scaled.map(([x, y]) => `${x},${1 - y}`).join(' ');
    return `0,1 ${map} 1,1`;
  }
}
