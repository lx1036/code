<?php


namespace Next\Config;


use Illuminate\Support\Arr;

class Repository
{
    protected $items = [];

    public function __construct($items = [])
    {
        $this->items = $items;
    }

    public function get($key, $default = null)
    {
        return Arr::get($this->items, $key, $default);
    }

    public function set($key, $value)
    {
        Arr::set($this->items, $key, $value);
    }
}