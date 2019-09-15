<?php


namespace Next\Foundation\Support;


use Next\Foundation\Container\Application;

class ServiceProvider
{
    /** @var Application  */
    protected $app;

    public function __construct(Application $app)
    {
        $this->app = $app;
    }
}