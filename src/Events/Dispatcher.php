<?php


namespace Next\Events;

/**
 * @see also https://github.com/thephpleague/event
 *
 */
class Dispatcher implements DispatcherInterface
{
    protected $listeners = [];

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
}