<?php

include __DIR__ . "/../vendor/autoload.php";


interface AInterface {}

class BClass {}

class AClass implements AInterface {
    public function __construct(BClass $bclass)
    {
    }
}

class CClass {
    public function __construct($id)
    {
    }
}

class DClass implements AInterface {

}

class EClass {
    public function __construct(AInterface $a_interface)
    {
    }
}

class FClass {
    public function __construct(AInterface $a_interface)
    {
    }
}

class NoConstructor {
}

class NestedConstructor {
    public $no_constructor;
    public $a_interface;

    public function __construct(NoConstructor $no_constructor, AInterface $a_interface)
    {
        $this->no_constructor = $no_constructor;
        $this->a_interface = $a_interface;
    }
}


/////////////////////////test Reflection////////////////////////////////////////
/*
$reflector = new ReflectionClass(NestedConstructor::class);
$constructor = $reflector->getConstructor();
$dependencies = $constructor->getParameters();

foreach ($dependencies as $dependency) {
    var_dump($dependency, $dependency->getClass());
}

*/
/////////////////////////test Reflection////////////////////////////////////////


/**
 * https://github.com/php-fig/fig-standards/blob/master/accepted/PSR-11-container.md
 */


$container = new \Next\Foundation\Container\Container();

// simple binding
$container->bind(AInterface::class, function ($app) {
    return new AClass($app->resolve(BClass::class));
});
// binding a singleton
$container->singleton(AInterface::class, function ($app) {
    return new AClass($app->resolve(BClass::class));
});
// binding a instance
$aclass = new AClass(new BClass());
$container->instance(AInterface::class, $aclass);
// binding primitives
$container->when(CClass::class)->needs('id')->give(1);
// binding interface to implementation
$container->bind(AInterface::class, AClass::class);
// context binding
$container->when(EClass::class)->needs(AInterface::class)->give(AClass::class);
$container->when(FClass::class)->needs(AInterface::class)->give(DClass::class);
// tag binding
// extend binding


$container->bind(AClass::class);
$container->bind(NoConstructor::class);
$container->bind(NestedConstructor::class);

//$object1 = new NoConstructor();
//$object2 = new NoConstructor();

$instance = $container->resolve(NoConstructor::class);
$instance2 = $container->resolve(NoConstructor::class);
$instance_nested = $container->resolve(NestedConstructor::class);
var_dump($instance, $instance_nested->no_constructor, $instance_nested->a_interface);
//var_dump($instance, $instance2, $instance2 === $instance, $object1 === $object2);





