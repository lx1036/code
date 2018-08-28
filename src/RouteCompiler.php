<?php


namespace Next;


class RouteCompiler
{
    protected $route;

    public function __construct(Route $route)
    {
        $this->route = $route;
    }

    public function compile(): CompiledRoute
    {
        $uri = preg_replace('/\{}/', '{$1}', $this->route->uri());

        if ('' !== $this->route->getHost()) {
            $pattern = $this->compilePattern($this->route, true);
        }


        return new CompiledRoute('');
    }

    private function compilePattern(Route $route, bool $is_host): array
    {

    }
}