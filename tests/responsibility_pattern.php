<?php

include __DIR__ . '/../vendor/autoload.php';

class Request {
    /** @var array */
    protected $payload;

    public function __construct(array $payload = [])
    {
        $this->payload = $payload;
    }

    public function getPayload(): array
    {
        return $this->payload;
    }

    public function setPayload(array $payload)
    {
        $this->payload = $payload;
    }
}

class Response {
    public function toJson(Request $request): string
    {
        return \GuzzleHttp\json_encode($request->getPayload());
    }
}

class Router {
    public function dispatch($request): string
    {
        return (new Response())->toJson($request);
    }
}

class Pipe1 {
    public function handle(Request $request, Closure $next)
    {
        $payload = array_merge($request->getPayload(), ['pipe1' => 'pipe1']);

        $request->setPayload($payload);

        return $next($request);
    }
}

class Pipe2 {
    public function handle(Request $request, Closure $next, $user, $type)
    {
        $payload = array_merge($request->getPayload(), [$user => $type]);

        $request->setPayload($payload);

        return $next($request);
    }
}

//$pipeline = new \Illuminate\Pipeline\Pipeline(new \Illuminate\Foundation\Application(__DIR__ . '/../'));

class Pipeline {
    protected $passable;

    protected $pipes;

    protected $method = 'handle';

    public function send($passable)
    {
        $this->passable = $passable;

        return $this;
    }

    public function through(array $pipes)
    {
        $this->pipes = $pipes;

        return $this;
    }

    public function then(Closure $destination)
    {
        $pipeline = array_reduce(array_reverse($this->pipes), $this->carry(), $destination);

        /**
         * 1. $stack = function($passable) use ($destination, $pipe2) {}
         * 2. $pipeline = function($passable) use ($stack, $pipe1) {}
         */

        return $pipeline($this->passable);
    }

    public function carry()
    {
        return function ($stack, $pipe) {
            return function ($passable) use ($stack, $pipe) {
                if (is_callable($pipe)) {
                    return $pipe($passable, $stack);
                } elseif (is_string($pipe)) {
                    [$name, $parameters] = $this->parsePipe($pipe);

                    $pipe = new $name;
                    $parameters = array_merge([$passable, $stack], $parameters);
                } elseif (is_object($pipe)) {
                    $parameters = [$passable, $stack];
                }

                $response = method_exists($pipe, $this->method) ? $pipe->{$this->method}(...$parameters)
                    : $pipe(...$parameters);

                return $response;
            };
        };
    }

    public function parsePipe(string $pipe): array
    {
        [$name, $parameters] = array_pad(explode(':', $pipe, 2), 2, []);

        if (is_string($parameters)) {
            $parameters = explode(',', $parameters);
        }

        return [$name, $parameters];
    }
}

$request = new Request(['username' => 'user@example.com', 'password' => 'password']);

$pipes = [
    Pipe1::class,
    Pipe2::class. ':' . implode(',', ['john', 'advisor']),
    function($passable, $next) {
        $payload = array_merge($passable->getPayload(), ['pipe3' => 'pipe3']);
        $passable->setPayload($payload);

        return $next($passable);
    },
];

$destination = function ($request) {
    return (new Router)->dispatch($request);
};

$response = (new Pipeline)->send($request)->through($pipes)->then($destination);

dump($response);

/**
 * test array_reduce()
 */
/*$pipes_test = ['b', 'c', 'd'];
$result = array_reduce($pipes_test, function ($initial, $pipe) {
    dump([$initial, $pipe]);
    return $initial . $pipe;
}, 'a');
dump($result);*/
