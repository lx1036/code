<?php


namespace Next\Foundation\Http\Requests;


class JsonRequest extends FormRequest
{
    /**
     * @return array [attribute => rule, ...]
     */
    public function rules(): array
    {
        assert(in_array($this->getMethod(), [static::METHOD_POST, static::METHOD_PUT, static::METHOD_PATCH], true));

        $controller = $this->route()->getController();

        foreach ($controller::RULES as $method => $rules) {
            if (in_array($this->getMethod(), explode(',', $method), true)) {
                return $rules;
            }
        }
    }
}