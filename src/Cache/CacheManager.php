<?php


namespace Next\Cache;

use Next\Cache\Drivers\ArrayDriver;
use Next\Cache\Drivers\RedisDriver;
use Next\Foundation\Container\Application;

/**
 * @see https://github.com/php-fig/fig-standards/blob/master/accepted/PSR-16-simple-cache.md
 *
 * @see https://symfony.com/doc/current/components/cache.html
 *
 */
class CacheManager
{
    /** @var array */
    protected $drivers;

    protected $app;

    public function __construct(Application $app)
    {
        $this->app = $app;
    }

    public function driver(string $name = null)
    {
        $name = $name ?: $this->getDefaultDriver();

        return $this->drivers[$name] ?? $this->drivers[$name] = $this->resolve($name);
    }

    public function getDefaultDriver(): string
    {
        $this->app['config']['cache.default'];
    }

    private function resolve($name): Driver
    {
        $config = $this->getDriverConfig($name);

        $driverMethod = 'create' . ucfirst($config['driver']) . 'Driver';

        if (method_exists($this, $driverMethod)) {
            return $this->{$driverMethod}($config);
        }

        throw new \InvalidArgumentException("Cache driver [$name] is not supported.");
    }

    private function getDriverConfig($name)
    {
        return $this->app['config']["cache.stores.{$name}"];
    }

    private function createApcDriver(): Driver
    {
        
    }

    private function createArrayDriver()
    {
        return new ArrayDriver();
    }

    private function createDatabaseDriver()
    {

    }

    private function createFileDriver()
    {

    }

    private function createNullDriver()
    {

    }

    private function createRedisDriver($config): Driver
    {
//        return new RedisDriver();
    }
}