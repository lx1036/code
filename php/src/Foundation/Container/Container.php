<?php


namespace Next\Foundation\Container;

use ArrayAccess;
use Closure;
use ReflectionClass;


/**
 * Bind/Resolve
 *
 * 1. bind abstract into Container, make concrete from Container(use ReflectionClass)
 * 2. alias abstract
 *
 *
 *
 */
class Container implements ArrayAccess, ContainerInterface
{
    protected static $instance;


    /**
     * @var [abstract => ['concrete' => concrete, 'singleton' => singleton], ...]
     */
    protected $bindings = [];

    protected $aliases = [];

    /**
     * Container's shared instances
     * @var array
     */
    protected $instances = [];

    protected static function setInstance(Container $container)
    {
        return static::$instance = $container;
    }

    /**
     * Bind singleton/or-not concrete
     *
     * @param string $abstract
     * @param string|\Closure|null $concrete
     * @param bool $singleton
     */
    public function bind($abstract, $concrete = null, bool $singleton = false)
    {
        // drop stale bindings
        unset($this->bindings[$abstract], $this->aliases[$abstract]);

        if (is_null($concrete)) {
            $concrete = $abstract;
        }

        // $concrete is classname, convert it into Closure
        if (! $concrete instanceof Closure) {
            $concrete = $this->getClosure($abstract, $concrete);
        }

        $this->bindings[$abstract] = compact('concrete', 'singleton');
    }

    /**
     * Bind a existing instance as shared into container.
     *
     * @param $abstract
     * @param $instance
     */
    public function instance($abstract, $instance)
    {
        $this->instances[$abstract] = $instance;
    }

    /**
     * 实现容器的单例模式
     *
     * @param      $abstract
     * @param null $concrete
     */
    public function singleton($abstract, $concrete = null)
    {
        $this->bind($abstract, $concrete, true);
    }

    /**
     * Resolve the given type from the container
     *
     * @param $abstract
     * @return mixed
     * @throws \Exception
     */
    public function resolve($abstract)
    {
        $abstract = $this->getAlias($abstract);

        if (isset($this->instances[$abstract])) {
            return $this->instances[$abstract];
        }

        /** @var Closure $concrete */
        $concrete = $this->bindings[$abstract]['concrete'];

        if ($concrete instanceof Closure) {
            return $this->instances[$abstract] = $concrete($this);
        } else {
            return $this->instantiate($concrete);
        }

        $object = $this->make($concrete);

        $this->fireAfterResolvingCallbacks($abstract, $object);

        return $object;
    }

    public function make($abstract)
    {
        return $this->resolve($abstract);
    }

    private function getClosure($abstract, $concrete): Closure
    {
        return function () use ($abstract, $concrete) {
            if ($abstract === $concrete) {
                return $this->instantiate($concrete);
            }

            return $this->resolve($concrete);
        };
    }

    private function getAlias($abstract)
    {
        return $this->aliases[$abstract] ?? $abstract;
    }

    /**
     * Instantiate a concrete instance of the given type.
     *
     * @param $concrete
     *
     * @return mixed
     * @throws \Exception
     */
    private function instantiate($concrete)
    {
        try {
            $reflector = new ReflectionClass($concrete);
        } catch (\ReflectionException $exception) {

        }


        if (! $reflector->isInstantiable()) {
            throw new \Exception("class [$concrete] Can\'t be instantiated.");
        }

        /** @var \ReflectionMethod $constructor */
        $constructor = $reflector->getConstructor();

        if (is_null($constructor)) {
            return new $concrete;
        }

        /** @var \ReflectionParameter[] $dependencies */
        $dependencies = $constructor->getParameters();
        $arguments = $this->resolveDependencies($dependencies);

        return $reflector->newInstanceArgs($arguments);
    }

    private function resolveClass(\ReflectionParameter $dependency)
    {
        return $this->resolve($dependency->getClass()->name);
    }

    private function resolveDependencies(array $dependencies): array
    {
        $arguments = [];

        foreach ($dependencies as $dependency) {
            $arguments[] = is_null($dependency->getClass()) ? $dependency : $this->resolveClass($dependency);
        }

        return $arguments;
    }

    public function offsetExists($offset)
    {
        // TODO: Implement offsetExists() method.
    }

    public function offsetGet($offset)
    {
        // TODO: Implement offsetGet() method.
    }

    public function offsetSet($offset, $value)
    {
        // TODO: Implement offsetSet() method.
    }

    public function offsetUnset($offset)
    {
        // TODO: Implement offsetUnset() method.
    }

    protected $afterResolvingCallbacks = [];

    public function afterResolving($abstract, \Closure $callback)
    {
        if (is_string($abstract)) {
            $abstract = $this->getAlias($abstract);
        }

        $this->afterResolvingCallbacks[$abstract][] = $callback;
    }

    private function fireAfterResolvingCallbacks($abstract, $object)
    {
        $callbacks = $this->afterResolvingCallbacks[$abstract];

        foreach ($callbacks as $callback) {
            $callback($object);
        }
    }
}
