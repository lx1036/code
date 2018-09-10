<?php


namespace Next\Routing\Validators;


use Next\Routing\Route;
use Symfony\Component\HttpFoundation\Request;

interface RequestValidatorInterface
{
    public function matches(Route $route, Request $request): bool ;
}