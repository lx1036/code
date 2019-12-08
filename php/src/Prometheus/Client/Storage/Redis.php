<?php


namespace Next\Prometheus\Client\Storage;


class Redis implements Adapter
{

    public static function setDefaultOptions(array $array)
    {
    }

    public function collect()
    {
        // TODO: Implement collect() method.
    }

    public function updateHistogram(array $data): void
    {
        // TODO: Implement updateHistogram() method.
    }

    public function updateGauge(array $data): void
    {
        // TODO: Implement updateGauge() method.
    }

    public function updateCounter(array $data): void
    {
        // TODO: Implement updateCounter() method.
    }
}
