<?php


namespace Next\Prometheus\Client;


use Next\Prometheus\Client\Storage\Adapter;

class Counter extends Collector
{
    const TYPE = 'counter';


    public function incBy($count, array $labels = []): void
    {
        $this->assertLabelsAreDefinedCorrectly($labels);

        $this->storageAdapter->updateCounter([
            'name' => $this->getName(),
            'help' => $this->getHelp(),
            'type' => $this->getType(),
            'labelNames' => $this->getLabelNames(),
            'labelValues' => $labels,
            'value' => $count,
            'command' => Adapter::COMMAND_INCREMENT_INTEGER,
        ]);
    }

    /**
     * @return string
     */
    public function getType(): string
    {
        return self::TYPE;
    }
}
