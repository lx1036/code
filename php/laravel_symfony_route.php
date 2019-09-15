<?php

include __DIR__ . "/vendor/autoload.php";

function AClosure() {

}


$routes = new \Symfony\Component\Routing\RouteCollection();
$route = new \Symfony\Component\Routing\Route('/foo', ['_controller' => 'AController']);

$routes->add('fooRoute', $route);
$matcher = new \Symfony\Component\Routing\Matcher\UrlMatcher($routes, new \Symfony\Component\Routing\RequestContext('/bar'));

//var_dump($matcher->match('/bar/foo'));



$lara_route = new \Illuminate\Routing\Route('GET', 'foo', ['uses' => function() {return 'bar';},]);
$lara_route2 = new \Illuminate\Routing\Route('POST', 'foo', ['uses' => function() {return 'bar';},]);
$lara_route3 = new \Illuminate\Routing\Route('DELETE', 'foo3', ['uses' => function() {return 'bar';}]);
$lara_routes = new \Illuminate\Routing\RouteCollection();
$lara_routes->add($lara_route);
$lara_routes->add($lara_route2);
$lara_routes->add($lara_route3);

$routes = $lara_routes->getRoutesByMethod();

var_dump($routes, $lara_routes->getRoutesByName());







//$lara_router = new \Illuminate\Routing\Router();
//$response = $lara_router->dispatch(new \Illuminate\Http\Request());
