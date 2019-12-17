<?php


namespace Next\Prometheus\Client\Storage;


use APCUIterator;
use Next\Prometheus\Client\MetricFamilySamples;
use RuntimeException;

class Apcu implements Adapter
{
    const PROMETHEUS_PREFIX = 'prom';

    public function __construct()
    {
    }

    public function collect(): array
    {
        $metrics = $this->collectHistograms();
        $metrics = array_merge($metrics, $this->collectGauges());
        $metrics = array_merge($metrics, $this->collectCounters());
        return $metrics;
    }

    public function updateHistogram(array $data): void
    {
        // Initialize the sum
        $sumKey = $this->histogramBucketValueKey($data, 'sum');
        $new = apcu_add($sumKey, $this->toInteger(0));
        // If sum does not exist, assume a new histogram and store the metadata
        if ($new) {
            apcu_store($this->metaKey($data), json_encode($this->metaData($data)));
        }

        // Atomically increment the sum
        // Taken from https://github.com/prometheus/client_golang/blob/66058aac3a83021948e5fb12f1f408ff556b9037/prometheus/value.go#L91
        $done = false;
        while (!$done) {
            $old = apcu_fetch($sumKey);
            $done = apcu_cas($sumKey, $old, $this->toInteger($this->fromInteger($old) + $data['value']));
        }
        // Figure out in which bucket the observation belongs
        $bucketToIncrease = '+Inf';
        foreach ($data['buckets'] as $bucket) {
            if ($data['value'] <= $bucket) {
                $bucketToIncrease = $bucket;
                break;
            }
        }

        // Initialize and increment the bucket
        apcu_add($this->histogramBucketValueKey($data, $bucketToIncrease), 0);
        apcu_inc($this->histogramBucketValueKey($data, $bucketToIncrease));
    }

    private function histogramBucketValueKey(array $data, $bucket): string
    {
        return implode(':', [
            self::PROMETHEUS_PREFIX,
            $data['type'],
            $data['name'],
            $this->encodeLabelValues($data['labelValues']),
            $bucket,
            'value',
        ]);
    }

    private function toInteger($val): int
    {
        return unpack('Q', pack('d', $val))[1];
    }

    public function updateGauge(array $data): void
    {
        $valueKey = $this->valueKey($data);
        if ($data['command'] == Adapter::COMMAND_SET) {
            apcu_store($valueKey, $this->toInteger($data['value']));
            apcu_store($this->metaKey($data), json_encode($this->metaData($data)));
        } else {
            $new = apcu_add($valueKey, $this->toInteger(0));
            if ($new) {
                apcu_store($this->metaKey($data), json_encode($this->metaData($data)));
            }
            // Taken from https://github.com/prometheus/client_golang/blob/66058aac3a83021948e5fb12f1f408ff556b9037/prometheus/value.go#L91
            $done = false;
            while (!$done) {
                $old = apcu_fetch($valueKey);
                $done = apcu_cas($valueKey, $old, $this->toInteger($this->fromInteger($old) + $data['value']));
            }
        }
    }

    public function updateCounter(array $data): void
    {
        $new = apcu_add($this->valueKey($data), 0);
        if ($new) {
            apcu_store($this->metaKey($data), json_encode($this->metaData($data)));
        }
        apcu_inc($this->valueKey($data), $data['value']);
    }

    private function metaData(array $data): array
    {
        $metricsMetaData = $data;
        unset($metricsMetaData['value']);
        unset($metricsMetaData['command']);
        unset($metricsMetaData['labelValues']);
        return $metricsMetaData;
    }

    private function metaKey(array $data): string
    {
        return implode(':', [self::PROMETHEUS_PREFIX, $data['type'], $data['name'], 'meta']);
    }

    private function valueKey(array $data): string
    {
        return implode(':', [
            self::PROMETHEUS_PREFIX,
            $data['type'],
            $data['name'],
            $this->encodeLabelValues($data['labelValues']),
            'value',
        ]);
    }

    private function encodeLabelValues(array $values): string
    {
        return base64_encode(\GuzzleHttp\json_encode($values));
    }

    private function collectHistograms()
    {
        $histograms = [];
        foreach (new APCUIterator('/^prom:histogram:.*:meta/') as $histogram) {
            $metaData = \GuzzleHttp\json_decode($histogram['value'], true);
            $data = [
                'name' => $metaData['name'],
                'help' => $metaData['help'],
                'type' => $metaData['type'],
                'labelNames' => $metaData['labelNames'],
                'buckets' => $metaData['buckets'],
            ];
            // Add the Inf bucket so we can compute it later on
            $data['buckets'][] = '+Inf';
            $histogramBuckets = [];
            foreach (new APCUIterator('/^prom:histogram:' . $metaData['name'] . ':.*:value/') as $value) {
                $parts = explode(':', $value['key']);
                $labelValues = $parts[3];
                $bucket = $parts[4];
                // Key by labelValues
                $histogramBuckets[$labelValues][$bucket] = $value['value'];
            }
            // Compute all buckets
            $labels = array_keys($histogramBuckets);
            sort($labels);
            foreach ($labels as $labelValues) {
                $acc = 0;
                $decodedLabelValues = $this->decodeLabelValues($labelValues);
                foreach ($data['buckets'] as $bucket) {
                    $bucket = (string)$bucket;
                    if (!isset($histogramBuckets[$labelValues][$bucket])) {
                        $data['samples'][] = [
                            'name' => $metaData['name'] . '_bucket',
                            'labelNames' => ['le'],
                            'labelValues' => array_merge($decodedLabelValues, [$bucket]),
                            'value' => $acc,
                        ];
                    } else {
                        $acc += $histogramBuckets[$labelValues][$bucket];
                        $data['samples'][] = [
                            'name' => $metaData['name'] . '_' . 'bucket',
                            'labelNames' => ['le'],
                            'labelValues' => array_merge($decodedLabelValues, [$bucket]),
                            'value' => $acc,
                        ];
                    }
                }
                // Add the count
                $data['samples'][] = [
                    'name' => $metaData['name'] . '_count',
                    'labelNames' => [],
                    'labelValues' => $decodedLabelValues,
                    'value' => $acc,
                ];
                // Add the sum
                $data['samples'][] = [
                    'name' => $metaData['name'] . '_sum',
                    'labelNames' => [],
                    'labelValues' => $decodedLabelValues,
                    'value' => $this->fromInteger($histogramBuckets[$labelValues]['sum']),
                ];
            }

            $histograms[] = $data;
        }

        return $histograms;
    }

    private function collectGauges(): array
    {
        $gauges = [];
        foreach (new APCUIterator('/^prom:gauge:.*:meta/') as $gauge) {
            $metaData = \GuzzleHttp\json_decode($gauge['value'], true);
            $data = [
                'name' => $metaData['name'],
                'help' => $metaData['help'],
                'type' => $metaData['type'],
                'labelNames' => $metaData['labelNames'],
            ];
            foreach (new APCUIterator('/^prom:gauge:' . $metaData['name'] . ':.*:value/') as $value) {
                $parts = explode(':', $value['key']);
                $labelValues = $parts[3];
                $data['samples'][] = [
                    'name' => $metaData['name'],
                    'labelNames' => [],
                    'labelValues' => $this->decodeLabelValues($labelValues),
                    'value' => $this->fromInteger($value['value']),
                ];
            }

            $this->sortSamples($data['samples']);
            $gauges[] = $data;
        }
        return $gauges;
    }

    /**
     * @return array
     */
    private function collectCounters(): array
    {
        $counters = [];

        foreach (new APCUIterator('/^prom:counter:.*:meta/') as $counter) {
            $metaData = \GuzzleHttp\json_decode($counter['value'], true);
            $data = [
                'name' => $metaData['name'],
                'help' => $metaData['help'],
                'type' => $metaData['type'],
                'labelNames' => $metaData['labelNames'],
            ];

            foreach (new APCUIterator('/^prom:counter:' . $metaData['name'] . ':.*:value/') as $value) {
                $parts = explode(':', $value['key']);
                $labelValues = $parts[3];
                $data['samples'][] = [
                    'name' => $metaData['name'],
                    'labelNames' => [],
                    'labelValues' => $this->decodeLabelValues($labelValues),
                    'value' => $value['value'],
                ];
            }

            $this->sortSamples($data['samples']);

            $counters[] = $data;
        }

        return $counters;
    }

    /**
     * @param string $values
     * @return array
     * @throws RuntimeException
     */
    private function decodeLabelValues($values): array
    {
        $json = base64_decode($values, true);
        if (false === $json) {
            throw new RuntimeException('Cannot base64 decode label values');
        }
        
        return \GuzzleHttp\json_decode($json, true);
    }

    /**
     * @param mixed $val
     * @return float
     */
    private function fromInteger($val): float
    {
        return unpack('d', pack('Q', $val))[1];
    }

    private function sortSamples(array &$samples): void
    {
        usort($samples, function ($a, $b) {
            return strcmp(implode("", $a['labelValues']), implode("", $b['labelValues']));
        });
    }

    public function flushAPC(): void
    {
        apcu_clear_cache();
    }
}
