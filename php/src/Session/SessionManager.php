<?php


namespace Next\Session;

use Next\Foundation\Container\Application;
use Next\Foundation\Support\Manager;

/**
 * Build the main objects
 *
 * @see http://php.net/manual/zh/class.sessionhandlerinterface.php
 * Handlers: Cache, Cookie, Database, File, Null
 *
 * @see https://symfony.com/doc/current/components/http_foundation/sessions.html
 * Symfony Session(symfony/http-foundation/Session) 使用的是 php 的 session
 */
class SessionManager extends Manager
{


    protected function createDatabaseDriver()
    {

    }

    protected function createFileDriver()
    {
        return $this->buildSession(new FileSessionHandler($this->app['filesystem'], $this->app['config']['session.file_path']));
    }

    protected function createRedisDriver()
    {

    }

    protected function buildSession($handler)
    {
        return new Store($this->app['config']['session.cookie'], $handler);
    }

    protected function getDefaultDriver()
    {
        return $this->app['config']['session.default'];
    }
}