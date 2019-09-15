<?php


namespace Next\Auth;


use Next\Foundation\Support\ServiceProvider;

class AuthServiceProvider extends ServiceProvider
{
    public function register()
    {
        $this->app->singleton('auth', function ($app) {
            return new AuthManager($app);
        });

        $this->app->singleton('auth.driver', function ($app) {
            $app['auth']->driver();
        });
    }
}