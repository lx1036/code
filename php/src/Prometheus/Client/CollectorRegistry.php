<?php


namespace Next\Prometheus\Client;


use Next\Prometheus\Client\Exception\MetricNotFoundException;
use Next\Prometheus\Client\Exception\MetricsRegistrationException;
use Next\Prometheus\Client\Storage\Adapter;
use Next\Prometheus\Client\Storage\Apcu;

class CollectorRegistry
{
    private Adapter $storageAdapter;

    /**
     * @var Counter[]
     */
    private array $counters = [];

    /**
     * @var Gauge[]
     */
    private array $gauges = [];

    /**
     * @var Histogram[]
     */
    private array $histograms = [];

    private static CollectorRegistry $defaultRegistry;

    public function __construct(Adapter $adapter)
    {
        $this->storageAdapter = $adapter;
    }

    public static function getDefault()
    {
        if (!self::$defaultRegistry) {
            return self::$defaultRegistry = new static(new Apcu());
        }
        return self::$defaultRegistry;
    }

    public function collect(): array
    {
        return $this->storageAdapter->collect();
    }

    /**
     * @param string $namespace
     * @param string $name
     * @param string $help
     * @param array $labels
     * @return Counter
     * @throws MetricsRegistrationException
     */
    public function registerCounter(string $namespace, string $name, string $help, array $labels = []): Counter
    {
        $metricIdentifier = self::metricIdentifier($namespace, $name);
        if (isset($this->counters[$metricIdentifier])) {
            throw new MetricsRegistrationException("Metric already registered");
        }
        $this->counters[$metricIdentifier] = new Counter(
            $this->storageAdapter,
            $namespace,
            $name,
            $help,
            $labels
        );

        return $this->counters[$metricIdentifier];
    }

    public function getOrRegisterCounter($namespace, $name, $help, $labels = []): Counter
    {
        try {
            $counter = $this->getCounter($namespace, $name);
        } catch (MetricNotFoundException $e) {
            $counter = $this->registerCounter($namespace, $name, $help, $labels);
        }

        return $counter;
    }

    private static function metricIdentifier($namespace, $name): string
    {
        return $namespace . ":" . $name;
    }

    /**
     * @param string $namespace
     * @param string $name
     * @param string $help
     * @param array $labels
     * @return Gauge
     * @throws MetricsRegistrationException
     */
    public function registerGauge(string $namespace, string $name, string $help, array $labels = []): Gauge
    {
        $metricIdentifier = self::metricIdentifier($namespace, $name);
        if (isset($this->gauges[$metricIdentifier])) {
            throw new MetricsRegistrationException("Metric already registered");
        }
        $this->gauges[$metricIdentifier] = new Gauge(
            $this->storageAdapter,
            $namespace,
            $name,
            $help,
            $labels
        );
        return $this->gauges[$metricIdentifier];
    }

    /**
     * @param $namespace
     * @param $name
     * @param $help
     * @param array $labels
     * @param null $buckets
     * @return Histogram
     * @throws MetricsRegistrationException
     */
    public function registerHistogram($namespace, $name, $help, $labels = [], $buckets = null): Histogram
    {
        $metricIdentifier = self::metricIdentifier($namespace, $name);
        if (isset($this->histograms[$metricIdentifier])) {
            throw new MetricsRegistrationException("Metric already registered");
        }
        $this->histograms[$metricIdentifier] = new Histogram(
            $this->storageAdapter,
            $namespace,
            $name,
            $help,
            $labels,
            $buckets
        );
        return $this->histograms[$metricIdentifier];
    }
}
