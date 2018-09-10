<?php


namespace Next\Routing;


use Next\Routing\Validators\HostValidator;
use Next\Routing\Validators\MethodValidator;
use Next\Routing\Validators\RequestValidatorInterface;
use Next\Routing\Validators\SchemaValidator;
use Next\Routing\Validators\UriValidator;
use Symfony\Component\HttpFoundation\Request;

class Route
{
    /**
     * @var CompiledRoute
     */
    public $compiled;

    /**
     * @var array
     */
    protected $methods;

    /**
     * @var string
     */
    protected $uri;

    /**
     * @var array
     */
    protected $action;


    protected static $validators = [
        MethodValidator::class,
        SchemaValidator::class,
        HostValidator::class,
        UriValidator::class,
    ];


    /**
     * Route constructor.
     * @param array|string $methods
     * @param string $uri
     * @param string|array|\Closure $action
     */
    public function __construct($methods, string $uri, $action)
    {
        $this->methods = $methods;
        $this->uri = $uri;
        $this->action = RouteAction::parse($this->uri, $action);
//        $this->action = $action;
    }


    /**
     * @param Request $request
     * @return bool
     */
    public function matches(Request $request): bool
    {
        $this->compileRoute();

        foreach (static::$validators as $validator) {
            /** @var RequestValidatorInterface $validator */
            if (! $validator->matches($this, $request)) {
                return false;
            }
        }

        return true;
    }

    protected function compileRoute(): CompiledRoute
    {
        if (!$this->compiled) {
            $this->compiled = (new RouteCompiler($this))->compile();
        }

        return $this->compiled;
    }

    public function getHost(): ?string
    {
        return isset($this->action['host']) ? str_replace(['https://', 'http://'], '', $this->action['host']) : null;
    }

    public function getUri(): string
    {
        return $this->uri;
    }

    public function bind(Request $request): Route
    {
    }

    public function getMethods(): array
    {
        return $this->methods;
    }

    public function getCompiled(): CompiledRoute
    {
        return $this->compiled;
    }

    public function run()
    {
    }
}