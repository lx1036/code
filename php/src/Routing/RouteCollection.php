<?php


namespace Next\Routing;

use Next\Routing\Exception\ResourceNotFoundException;
use Symfony\Component\HttpFoundation\Request;
use Traversable;

/**
 * @depends illuminate/routing symfony/routing nikic/fast-route
 */
class RouteCollection implements \IteratorAggregate, \Countable
{
    /**
     * @var array
     */
    protected $routes = [];

    public function getIterator(): Traversable
    {
        return new \ArrayIterator($this->routes);
    }

    public function count(): int
    {
        return count($this->routes);
    }

    public function match(Request $request): Route
    {
        /** @var Route[] $routes */
        $routes = $this->get($request->getMethod());

        /** @var Route $route */
        $route = $this->matchRoutes($routes, $request);

        if (!is_null($route)) {
            return $route->bind($request);
        }

        throw new ResourceNotFoundException();
    }

    /**
     * @param Route[] $routes
     * @param Request $request
     * @return Route|null
     */
    protected function matchRoutes(array $routes, Request $request): ?Route
    {
        foreach ($routes as $route) {
            /** @var Route $route */
            if ($route->matches($request)) {
                return $route;
            }
        }

        return null;
    }

    public function add(Route $route): Route
    {
        $host_uri = $route->getHost() . $route->getUri();

        foreach ($route->getMethods() as $method) {
            $this->routes[$method][$host_uri] = $route;
        }

        return $route;
    }

    public function get($method = null)
    {
        return is_null($method) ? $this->routes : ($this->routes[$method] ?? []);
    }
}