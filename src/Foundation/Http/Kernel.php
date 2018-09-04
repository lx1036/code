<?php


namespace Next\Foundation\Http;


use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

class Kernel
{
    public function __construct()
    {
    }

    public function handle(Request $request): Response
    {
        try {
            $response = $this->sendRequest($request);
        } catch (\Exception $exception) {

        }

        return $response;
    }

    /**
     * 1. load environment variables
     * 2. load configuration
     * 3.
     */
    public function bootstrap()
    {

    }

    private function sendRequest(Request $request)
    {
        $this->bootstrap();


    }
}