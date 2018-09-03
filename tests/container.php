<?php

include __DIR__ . "/../vendor/autoload.php";


interface AInterface {}

class AClass implements AInterface {

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











$container = new \Next\Foundation\Container\Container();
$container->bind(AClass::class);
$container->bind(NoConstructor::class);
$container->bind(NestedConstructor::class);
$container->bind(AInterface::class, AClass::class);

//$object1 = new NoConstructor();
//$object2 = new NoConstructor();

$instance = $container->make(NoConstructor::class);
$instance2 = $container->make(NoConstructor::class);
$instance_nested = $container->make(NestedConstructor::class);
var_dump($instance, $instance_nested->no_constructor, $instance_nested->a_interface);
//var_dump($instance, $instance2, $instance2 === $instance, $object1 === $object2);





