<?php


namespace Next\Validators;


use Next\Route;
use Symfony\Component\HttpFoundation\Request;

class SchemaValidator implements RequestValidatorInterface
{

    public function matches(Route $route, Request $request): bool
    {
        return true;
    }
}