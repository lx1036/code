<?php


namespace Next\Foundation\Support;


use Next\Foundation\Container\Application;

class Manager
{
    protected $app;

    /**
     * The registered custom driver creators.
     *
     * @var array
     */
    protected $customCreators = [];

    protected $drivers = [];


    public function __construct(Application $app)
    {
        $this->app = $app;
    }

    public function driver($name = null)
    {
        $name = $name ?: $this->getDefaultDriver();

        if (!isset($this->drivers[$name])) {
            $this->drivers[$name] = $this->createDriver($name);
        }

        return $this->drivers[$name];
    }

    abstract protected function getDefaultDriver();

    private function createDriver($name)
    {

    }
}