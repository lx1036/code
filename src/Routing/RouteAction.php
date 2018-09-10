<?php


namespace Next\Routing;


class RouteAction
{

    public static function parse(string $uri, $action): array
    {
        if (is_callable($action)) {
            return ['uses' => $action];
        } elseif (is_array($action) && isset($action['uses']) && is_string($action['uses'])) {
            return $action;
        }
    }
}