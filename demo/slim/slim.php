<?php

include __DIR__ . "/../../vendor/autoload.php";

use Psr\Http\Message\ServerRequestInterface as Request;
use Psr\Http\Message\ResponseInterface as Response;
use Slim\App;

//var_dump('asdfsadf');die();
$app = new App([
    'settings'                 => [
        'displayErrorDetails' => true,
    ],
]);
//$app->get('/hello/{name}', function (Request $request, Response $response, array $args) {
////    var_dump($request->getAttributes());
//
//    var_dump('asdfasfdsadf');
//    $response->getBody()->write('Hello, ' . $args['name']);
//
//    return $response;
//});
$app->get('/hello', function (Request $request, Response $response) {
    $response->getBody()->write('Hello, asdf');

    return $response;
});

//var_dump('asdf');
$app->run();
