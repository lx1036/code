<?php


namespace Next\Prometheus\Client;


use Next\Prometheus\Client\Storage\Adapter;

class Gauge extends Collector
{
    const TYPE = 'gauge';

    public function getType()
    {
        return self::TYPE;
    }

    public function set(float $value, array $labels = []): void
    {
        $this->assertLabelsAreDefinedCorrectly($labels);

        $this->storageAdapter->updateGauge(
            [
                'name' => $this->getName(),
                'help' => $this->getHelp(),
                'type' => $this->getType(),
                'labelNames' => $this->getLabelNames(),
                'labelValues' => $labels,
                'value' => $value,
                'command' => Adapter::COMMAND_SET,
            ]
        );
    }
}
