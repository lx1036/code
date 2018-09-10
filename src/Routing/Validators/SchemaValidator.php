<?php


namespace Next\Routing\Validators;


use Next\Routing\Route;
use Symfony\Component\HttpFoundation\Request;

class SchemaValidator implements RequestValidatorInterface
{

    public function matches(Route $route, Request $request): bool
    {
        return true;
    }
}