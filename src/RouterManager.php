<?php


namespace Next;


use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

class RouterManager
{
    /**
     * The current request being dispatched.
     *
     * @var \Symfony\Component\HttpFoundation\Request
     */
    protected $currentRequest;

    /**
     * @var RouteCollection
     */
    protected $routes;


    public function __construct()
    {
        $this->routes = new RouteCollection();
    }

    public function dispatch(Request $request): Response
    {
        return $this->runRoute($request, $this->findRoute($request));
    }

    protected function runRoute(Request $request, Route $route): Response
    {
        return $route->run();
    }

    /**
     * Match the route with current request from route collection.
     *
     * @param Request $request
     * @return Route
     */
    protected function findRoute(Request $request): Route
    {
        $route = $this->routes->match($request);


        return $route;
    }

    /**
     * @param string $uri
     * @param string|array|\Closure $action
     * @return Route
     */
    public function get(string $uri, $action): Route
    {
        return $this->addRoute(['GET'], $uri, $action);
    }

    /**
     * @param array $methods
     * @param string $uri
     * @param string|array|\Closure $action
     * @return Route
     */
    public function addRoute(array $methods, string $uri, $action): Route
    {
        return $this->routes->add($this->createRoute($methods, $uri, $action));
    }

    /**
     * @param array $methods
     * @param string $uri
     * @param string|array|\Closure $action
     * @return Route
     */
    private function createRoute(array $methods, string $uri, $action): Route
    {
        if ($this->isController($action)) {
            $action = $this->convertToController($action);
        }

        $route = new Route($methods, $uri, $action);

        return $route;
    }

    private function isController($action)
    {
        if (! $action instanceof \Closure) {
            return is_string($action) || (isset($action['uses']) && is_string($action['uses']));
        }

        return false;
    }

    private function convertToController($action): array
    {
        if (is_string($action)) {
            $action = ['uses' => $action];
        }

        return $action;
    }

    public function getRoutes(): RouteCollection
    {
        return $this->routes;
    }
}