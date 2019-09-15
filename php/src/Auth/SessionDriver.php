<?php


namespace Next\Auth;


class SessionDriver implements Driver
{
    protected $provider;

    public function __construct(UserProvider $provider)
    {
        $this->provider = $provider;
    }

    public function user()
    {
        // TODO: Implement user() method.
    }
}