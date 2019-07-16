<?php
include __DIR__ . '/../../../../vendor/autoload.php';


class Singleton {
    private static $singleton = null;

    public static function getInstance() {
        if (!static::$singleton) {
            static::$singleton = new self();
        }

        return static::$singleton;
    }
}

$instance1 = Singleton::getInstance();
$instance2 = Singleton::getInstance();

var_dump($instance1 === $instance2);


/**
 * Container 的单例模式
 */
$container = new \Next\Foundation\Container\Container();
$container->singleton(Singleton::class);
$singleton1 = $container->resolve(Singleton::class);
$singleton2 = $container->resolve(Singleton::class);
var_dump($singleton1 === $singleton2);
