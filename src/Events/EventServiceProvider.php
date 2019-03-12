<?php


namespace Next\Events;


use Next\Foundation\Support\ServiceProvider;
use Next\Queue\QueueInterface;

class EventServiceProvider extends ServiceProvider
{
    public function register()
    {
        $this->app->singleton('events', function ($app) {
            return (new Dispatcher($app))->setQueueResolver(function () use ($app) {
                return $app->make(QueueInterface::class);
            });
        });
    }
}