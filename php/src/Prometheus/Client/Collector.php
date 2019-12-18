<?php


namespace Next\Prometheus\Client;


use Next\Prometheus\Client\Storage\Adapter;
use InvalidArgumentException;

abstract class Collector
{
    protected Adapter $storageAdapter;

    const RE_METRIC_LABEL_NAME = '/^[a-zA-Z_:][a-zA-Z0-9_:]*$/';
    protected string $name;
    protected string $help;
    protected array $labels;
    /**
     * @param Adapter $storageAdapter
     * @param string $namespace
     * @param string $name
     * @param string $help
     * @param array $labels
     */
    public function __construct(Adapter $storageAdapter, $namespace, $name, $help, $labels = [])
    {
        $this->storageAdapter = $storageAdapter;
        $metricName = ($namespace ? $namespace . '_' : '') . $name;
        if (!preg_match(self::RE_METRIC_LABEL_NAME, $metricName)) {
            throw new InvalidArgumentException("Invalid metric name: '" . $metricName . "'");
        }
        $this->name = $metricName;
        $this->help = $help;
        foreach ($labels as $label) {
            if (!preg_match(self::RE_METRIC_LABEL_NAME, $label)) {
                throw new InvalidArgumentException("Invalid label name: '" . $label . "'");
            }
        }
        $this->labels = $labels;
    }

    protected function assertLabelsAreDefinedCorrectly(array $labels): void
    {
        if (count($labels) != count($this->labels)) {
            throw new InvalidArgumentException(sprintf('Labels are not defined correctly: ', print_r($labels, true)));
        }
    }

    abstract public function getType();
    /**
     * @return string
     */
    public function getName(): string
    {
        return $this->name;
    }

    public function getLabelNames(): array
    {
        return $this->labels;
    }
    /**
     * @return string
     */
    public function getHelp(): string
    {
        return $this->help;
    }
}
