<?php


namespace Next\Validation\Exceptions;


use Next\Validation\Validator;
use Throwable;

class ValidationException extends \Exception
{
    /** @var Validator */
    protected $validator;

    public function __construct(Validator $validator, string $message = "", int $code = 0, Throwable $previous = null)
    {
        parent::__construct($message, $code, $previous);

        $this->validator = $validator;
    }
}