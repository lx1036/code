<?php


namespace Next\Cache\Drivers;


use Next\Cache\Driver;
use Next\Redis\RedisInterface;

class RedisDriver implements Driver
{
    public function __construct(RedisInterface $redis)
    {
    }
}