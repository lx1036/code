<?php


namespace Next\Foundation\Http;


use Next\Foundation\Bootstrap\HandleExceptions;
use Next\Foundation\Container\Application;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

class Kernel
{
    protected $app;

    public function __construct(Application $app)
    {
        $this->app = $app;
    }

    public function handle(Request $request): Response
    {
        try {
            $response = $this->sendRequest($request);
        } catch (\Exception $exception) {

        }

        return $response;
    }

    protected $bootstrappers = [
        HandleExceptions::class
    ];

    /**
     * 1. load environment variables
     * 2. load configuration
     * 3. exception handler
     */
    public function bootstrap()
    {
        foreach ($this->bootstrappers as $bootstrapper) {
            $this->app->make($bootstrapper)->bootstrap($this->app);
        }
    }

    private function sendRequest(Request $request)
    {
        $this->bootstrap();


    }
}