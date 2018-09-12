<?php


namespace Next\Foundation\Container;


use Next\Events\EventServiceProvider;
use Next\Foundation\Support\ServiceProvider;
use Next\Log\LogServiceProvider;
use Next\Routing\RoutingServiceProvider;

class Application extends Container
{
    protected $service_providers = [];

    protected $booted = false;


    public function __construct($path = null)
    {
        /**
         * bind basic instances into container
         * e.g. 'router', 'log'
         */

        static::setInstance($this);

        $this->instance('app', $this);
        $this->instance(Container::class, $this);

        /**
         * register basic service providers
         */
        $this->register(new EventServiceProvider($this));
        $this->register(new LogServiceProvider($this));
        $this->register(new RoutingServiceProvider($this));
    }

    /**
     * @param \Next\Foundation\Support\ServiceProvider|string $provider
     * @param bool $force
     * @return ServiceProvider|string
     */
    public function register($provider, $force = false)
    {
        if (($registered = $this->getProvider($provider)) && ! $force) {
            return $registered;
        }

        if (is_string($provider)) {
            $provider = new $provider($this);
        }

        $this->service_providers[] = $provider;

        if ($this->booted) {
            if (method_exists($provider, 'boot')) {
                $this->call([$provider, 'boot']);
            }
        }

        return $provider;
    }

    /**
     * @param $provider
     * @return ServiceProvider
     */
    private function getProvider($provider): ServiceProvider
    {
        $name = is_string($provider) ? $provider : get_class($provider);

        $providers = array_filter($this->service_providers, function ($value) use ($name): bool {
            return $value instanceof $name;
        });

        return array_values($providers)[0];
    }

    public function runInConsole(): bool
    {
        return php_sapi_name() === 'cli';
    }
}