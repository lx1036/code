<?php


namespace Next\Cache\Drivers;


class ArrayDriver
{
    protected $storage = [];

    public function get($key)
    {
        return $this->storage[$key] ?? null;
    }

    public function put($key, $value, $minutes)
    {
        $this->storage[$key] = $value;
    }

    public function forget($key)
    {
        unset($this->storage[$key]);

        return true;
    }

    public function flush()
    {
        $this->storage = [];

        return true;
    }

    public function increment($key, $value = 1)
    {
        $this->storage[$key] = ! isset($this->storage[$key])
            ? $value : ((int) $this->storage[$key]) + $value;

        return $this->storage[$key];
    }

    public function decrement($key, $value = 1)
    {
        return $this->increment($key, $value * -1);
    }
}