<?php


namespace Next\Routing;


use Next\Foundation\Support\ServiceProvider;

class RoutingServiceProvider extends ServiceProvider
{
    public function register()
    {
        $this->app->singleton('router', function ($app) {
            return new RouterManager($app['events'], $app);
        });
    }
}