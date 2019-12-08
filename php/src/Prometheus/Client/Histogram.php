<?php


namespace Next\Prometheus\Client;


use Next\Prometheus\Client\Storage\Adapter;
use InvalidArgumentException;

class Histogram extends Collector
{
    /**
     * @var int[]
     */
    private $buckets;

    const TYPE = 'histogram';

    public function __construct(Adapter $storageAdapter, $namespace, $name, $help, $labels = [], $buckets = null)
    {
        foreach ($labels as $label) {
            if ($label === 'le') {
                throw new InvalidArgumentException("Histogram cannot have a label named 'le'.");
            }
        }

        parent::__construct($storageAdapter, $namespace, $name, $help, $labels);

        $this->buckets = $this->getBuckets($buckets);
    }

    private function getBuckets($buckets): array {
        if (null === $buckets) {
            $buckets = self::getDefaultBuckets();
        }
        if (0 == count($buckets)) {
            throw new InvalidArgumentException("Histogram must have at least one bucket.");
        }
        for ($i = 0; $i < count($buckets) - 1; $i++) {
            if ($buckets[$i] >= $buckets[$i + 1]) {
                throw new InvalidArgumentException(
                    "Histogram buckets must be in increasing order: " .
                    $buckets[$i] . " >= " . $buckets[$i + 1]
                );
            }
        }

        return $buckets;
    }

    public function getType(): string
    {
        return self::TYPE;
    }

    public static function getDefaultBuckets(): array
    {
        return [
            0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0, 7.5, 10.0,
        ];
    }

    public function observe(float $value, array $labels = []): void
    {
        $this->assertLabelsAreDefinedCorrectly($labels);
        $this->storageAdapter->updateHistogram(
            [
                'value' => $value,
                'name' => $this->getName(),
                'help' => $this->getHelp(),
                'type' => $this->getType(),
                'labelNames' => $this->getLabelNames(),
                'labelValues' => $labels,
                'buckets' => $this->buckets,
            ]
        );
    }
}
