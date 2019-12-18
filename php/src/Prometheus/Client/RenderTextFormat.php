<?php


namespace Next\Prometheus\Client;


class RenderTextFormat
{
    const MIME_TYPE = 'text/plain; version=0.0.4';

    public function render(array $metrics)
    {
        /*usort($metrics, function (MetricFamilySamples $a, MetricFamilySamples $b) {
            return strcmp($a->getName(), $b->getName());
        });*/
        usort($metrics, function (array $a, array $b) {
            return strcmp($a['name'], $b['name']);
        });
        $lines = [];
        foreach ($metrics as $metric) {
            $lines[] = "# HELP " . $metric['name'] . " {$metric['help']}";
            $lines[] = "# TYPE " . $metric['name'] . " {$metric['type']}";
            foreach ($metric['samples'] as $sample) {
                $lines[] = $this->renderSample($metric, $sample);
            }
        }
        return implode("\n", $lines) . "\n";
    }

    private function renderSample(array $metric, array $sample): string
    {
        $escapedLabels = [];
        $labelNames = $metric['labelNames'];
        if (!empty($metric['labelNames']) || !empty($sample['labelNames'])) {
            $labels = array_combine(array_merge($labelNames, $sample['labelNames']), $sample['labelValues']);
            foreach ($labels as $labelName => $labelValue) {
                $escapedLabels[] = $labelName . '="' . $this->escapeLabelValue($labelValue) . '"';
            }
            return $sample['name'] . '{' . implode(',', $escapedLabels) . '} ' . $sample['value'];
        }
        return $sample['name'] . ' ' . $sample['value'];
    }

    private function escapeLabelValue($v): string
    {
        $v = str_replace("\\", "\\\\", $v);
        $v = str_replace("\n", "\\n", $v);
        $v = str_replace("\"", "\\\"", $v);
        return $v;
    }
}
