<?php

$cache = new \Next\Cache\CacheManager(new \Next\Foundation\Container\Application());

$cache->driver('array')->put('key1', 'value1');

var_dump($cache->driver('array')->get('key1'));