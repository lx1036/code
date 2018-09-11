<?php


namespace Next\Foundation\Container;


use Next\Foundation\Support\ServiceProvider;

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

        $providers =  array_filter($this->service_providers, function ($value) use ($name): bool {
            return $value instanceof $name;
        });

        return array_values($providers)[0];
    }


}