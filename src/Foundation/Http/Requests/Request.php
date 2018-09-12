<?php


namespace Next\Foundation\Http\Requests;

use Closure;

class Request
{
    /** @var Closure */
    protected $route_resolver;

    protected $query_params;
    protected $form_data;
    protected $cookies;
    protected $files;
    protected $server;

    const METHOD_POST = 'POST';
    const METHOD_PUT = 'PUT';
    const METHOD_PATCH = 'PATCH';

    /**
     * Request constructor.
     *
     * @param array $query_params   The $_GET parameters
     * @see http://www.php.net/manual/zh/reserved.variables.post.php
     * @param array $form_data      The $_POST parameters, Content-Type is application/x-www-form-urlencoded or multipart/form-data
     * @param array $cookies        The $_COOKIE parameters
     * @param array $files          The $_FILES parameters
     * @param array $server         The $_SERVER parameters http://php.net/manual/zh/reserved.variables.server.php
     */
    public function __construct(array $query_params = [], array $form_data = [], array $cookies = [], array $files = [], array $server = [])
    {
    }


    public function setRouteResolver(Closure $callback)
    {
        $this->route_resolver = $callback;

        return $this;
    }

    public function getRouteResolver(): Closure
    {
        return $this->route_resolver ?: function() {};
    }

    protected $user_resolver;

    public function setUserResolver(Closure $callback)
    {
        $this->user_resolver = $callback;

        return $this;
    }

    public function getUserResolver()
    {
        return $this->user_resolver ?: function() {};
    }


    /**
     * @param null $param
     *
     * @return mixed|\Next\Routing\Route
     */
    public function route($param = null)
    {
        /** @var \Next\Routing\Route $route */
        $route = call_user_func($this->getRouteResolver());

        if (is_null($param)) {
            return $route;
        }

        return $route->parameter($param);
    }

    public function user($guard = null)
    {
        return call_user_func($this->getUserResolver(), $guard);
    }
}