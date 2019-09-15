<?php


namespace Next\Routing\Validators;


use Next\Routing\Route;
use Symfony\Component\HttpFoundation\Request;

class UriValidator implements RequestValidatorInterface
{

    public function matches(Route $route, Request $request): bool
    {
        $path = rtrim($request->getPathInfo(), '/');

        return (bool) preg_match($route->getCompiled()->getUriRegex(), rawurldecode($path));
    }
}