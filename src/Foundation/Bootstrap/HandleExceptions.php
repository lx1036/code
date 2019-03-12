<?php


namespace Next\Foundation\Bootstrap;


use Next\Foundation\Container\Application;
use ErrorException;
use Next\Foundation\Exceptions\ExceptionHandlerInterface;

/**
 * Register custom error and exception handler
 *
 */
class HandleExceptions
{
    protected $app;

    public function bootstrap(Application $app)
    {
        $this->app = $app;
        
        error_reporting(-1); // report all errors/exceptions

        set_error_handler([$this, 'handleError']);
        set_exception_handler([$this, 'handleException']);
    }

    public function handleError($level, $message, $file = '', $line = 0)
    {
        /**
         *  E_ERROR、 E_PARSE、 E_CORE_ERROR、 E_CORE_WARNING、 E_COMPILE_ERROR、 E_COMPILE_WARNING error level
         * is not handled by custom error handler
         */
        if (error_reporting() & $level) { // error code is included in error_reporting
            throw new ErrorException($message,0, $level, $file, $line);
        }
    }

    /**
     * report exception to external service like Sentry,
     * render the exception as response to send back to frontend
     *
     * @param \Throwable $e
     */
    public function handleException(\Throwable $e)
    {
        try {
            $this->getExceptionHandler()->report($e);
        } catch (\Exception $e) {

        }

        if ($this->app->runInConsole()) { // cli
            $this->renderForConsole($e);
        } else { // web
            $this->renderForWeb($e);
        }
    }

    private function getExceptionHandler()
    {
        return $this->app->make(ExceptionHandlerInterface::class);
    }

    private function renderForConsole($e)
    {
        // TODO: exception for cli
        $this->getExceptionHandler()->renderForConsole();
    }

    private function renderForWeb($e)
    {
        $this->getExceptionHandler()->render($this->app['request'], $e)->send();
    }
}