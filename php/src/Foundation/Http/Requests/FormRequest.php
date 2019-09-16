<?php


namespace Next\Foundation\Http\Requests;


use Next\Foundation\Http\Contracts\ValidatesWhenResolved;
use Next\Validation\Exceptions\ValidationException;
use Next\Validation\Validator;

class FormRequest extends Request implements ValidatesWhenResolved
{

    public function validateResolved(): void
    {
        $this->prepareForValidation();

        $validator = $this->getValidator();

        if ($validator->fails()) {
            throw new ValidationException($validator);
        }
    }

    private function prepareForValidation()
    {
    }

    private function getValidator(): Validator
    {
        // TODO: 2018-09-11 xiang.liu@rightcapital.com
    }
}