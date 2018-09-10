<?php


namespace Next\Routing\Validators;


use Next\Routing\Route;
use Symfony\Component\HttpFoundation\Request;

class MethodValidator implements RequestValidatorInterface
{

    public function matches(Route $route, Request $request): bool
    {
        return in_array($request->getMethod(), $route->getMethods(), true);
    }
}