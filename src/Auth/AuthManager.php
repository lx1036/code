<?php


namespace Next\Auth;


use Next\Foundation\Container\Application;

/**
 * @link https://symfony.com/doc/current/components/security/authentication.html
 *
 */
class AuthManager
{
    /** @var \Closure */
    protected $userResolver;

    protected $app;

    public function __construct(Application $app)
    {
        $this->app = $app;

        $this->userResolver = function ($driver) {
            return $this->driver($driver)->user();
        };
    }

    protected $drivers = [];

    public function driver(string $name): Driver
    {
        $name = $name ?: $this->getDefaultDriver();

        return $this->drivers[$name] ?? $this->drivers[$name] = $this->resolve($name);
    }

    protected function resolve($name)
    {
        $config = $this->getDriverConfig($name);

        $driverMethod = 'create' . ucfirst($config['driver']) . 'Driver';

        if (method_exists($this, $driverMethod)) {
            return $this->{$driverMethod}($name, $this->getUserProviderConfiguration($config));
        }

        throw new \InvalidArgumentException("Auth driver [$name] is not supported.");
    }

    public function createSessionDriver($name, array $config): Driver
    {
        $provider = $this->createUserProvider($config);

        $driver = new SessionDriver($provider);

        return $driver;
    }

    public function createUserProvider(array $config): UserProvider
    {
        switch ($driver = $config['driver']) {
            case 'database':
                return new DatabaseUserProvider();
            default:
                throw new \InvalidArgumentException("Auth user provider [$driver] is not supported.");
        }
    }

    private function getDefaultDriver()
    {
        return $this->app['config']['auth.default.driver'];
    }

    private function getDriverConfig($name)
    {
        return $this->app['config']["auth.drivers.{$name}"];
    }

    private function getUserProviderConfiguration(array $config)
    {
        return $this->app['config']["auth.providers.{$config['provider']}"];
    }


}