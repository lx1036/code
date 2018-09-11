<?php

include __DIR__ . "/../vendor/autoload.php";


$app = new \Next\Foundation\Container\Application();

$app->register(\Next\Foundation\FoundationServiceProvider::class);

$app->singleton(\Next\Foundation\Http\KernelInterface::class, \Next\Foundation\Http\Kernel::class);


/** @var \Next\Foundation\Http\Kernel $kernel */
$kernel = $app->make(\Next\Foundation\Http\KernelInterface::class);
$response = $kernel->handle(\Symfony\Component\HttpFoundation\Request::create('/foo',
    \Symfony\Component\HttpFoundation\Request::METHOD_GET));

var_dump($response->getContent());
