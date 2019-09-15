<?php

include __DIR__ . "/../vendor/autoload.php";

/**
 * https://symfony.com/doc/current/components/dependency_injection.html
 */

class Mail {
    protected $transport;

    public function __construct($transport)
    {
        $this->transport = $transport;
    }
}

$container_builder = new \Symfony\Component\DependencyInjection\ContainerBuilder();
$container_builder->register('mailer', Mail::class)->addArgument('sendmail');
$mailer = $container_builder->get('mailer');

var_dump($mailer);


