<?php


namespace Next\Routing\Validators;


use Next\Routing\Route;
use Symfony\Component\HttpFoundation\Request;

class HostValidator implements RequestValidatorInterface
{

    public function matches(Route $route, Request $request): bool
    {
        if (is_null($route->getCompiled()->getHostRegex())) {
            return true;
        }

        return (bool) preg_match($route->getCompiled()->getHostRegex(), $request->getHost());
    }
}