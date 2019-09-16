<?php


namespace Next\Events;


interface DispatcherInterface
{
    public function listen($events, $listener);

    public function fire($event, $payload);
}