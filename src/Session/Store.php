<?php


namespace Next\Session;

use SessionHandlerInterface;

/**
 * Do the main jobs
 */
class Store
{
    protected $name;

    protected $handler;

    public function __construct($name, SessionHandlerInterface $handler)
    {
        $this->name = $name;
        $this->handler = $handler;
    }
}