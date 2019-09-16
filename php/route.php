<?php

include __DIR__ . '/vendor/autoload.php';

use \Symfony\Component\HttpFoundation\Response;
use \Symfony\Component\HttpFoundation\Request;

class AClass {
    public function show(Request $request): Response
    {
        $id = $request->get('id');

        $content = "people[$id]";

        return new Response($content);
    }

    public function getUser(): Response
    {
        return new Response('user[1]');
    }
}

$router = new \Next\RouterManager();

$router->get('password', function () {
    return 'password';
});
$router->get('username/:id', AClass::class . '@show');
$router->get('user', ['uses' => AClass::class . '@getUser']);


//var_dump($router->getRoutes()->get());

$response = $router->dispatch(Request::create('password', 'GET'));

//var_dump($response->getContent());




