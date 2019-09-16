<?php


namespace Next\Events;

use Next\Foundation\Container\Container;
use Closure;

/**
 * @see also https://github.com/thephpleague/event
 *
 */
class Dispatcher implements DispatcherInterface
{
    protected $listeners = [];

    protected $container;

    public function __construct(Container $container)
    {
        $this->container = $container;
    }

    /**
     * @param array|string $events
     * @param string|\Closure $listener
     */
    public function listen($events, $listener)
    {
        foreach ((array)$events as $event) {
            $this->listeners[$event][] = $this->createListener($listener);
        }
        
    }

    public function fire($event, $payload)
    {
        $responses = [];
        $listeners = $this->listeners[$event];

        foreach ($listeners as $listener) {
            $response = $listener($payload);

            if ($response === false) {
                break;
            }

            $responses[] = $response;
        }

        return $responses;
    }

    /**
     * @param string|\Closure $listener
     * @return \Closure
     */
    private function createListener($listener): \Closure
    {

        if (is_string($listener)) {
            return function ($payload) use ($listener) {
                return call_user_func([$listener, 'handle'], $payload);
            };
        }

        return $listener;
    }

    protected $queueResolver;

    public function setQueueResolver(Closure $callback): Dispatcher
    {
        $this->queueResolver = $callback;

        return $this;
    }

    public function getQueueResolver()
    {
        return $this->queueResolver ?: function () {};
    }
}