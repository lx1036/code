<?php


namespace Next\Foundation;


use Next\Foundation\Http\Contracts\ValidatesWhenResolved;
use Next\Foundation\Support\ServiceProvider;

class FoundationServiceProvider extends ServiceProvider
{
    public function register()
    {

    }

    public function boot()
    {
        $this->app->afterResolving(ValidatesWhenResolved::class, function (ValidatesWhenResolved $resolved) {
            $resolved->validate();
        });
    }
}