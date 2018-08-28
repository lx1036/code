<?php


namespace Next\Validators;


use Next\Route;
use Symfony\Component\HttpFoundation\Request;

interface RequestValidatorInterface
{
    public function matches(Route $route, Request $request): bool ;
}