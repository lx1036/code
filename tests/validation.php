<?php

include __DIR__ . "/../vendor/autoload.php";


$translator = new \Illuminate\Translation\Translator(new \Illuminate\Translation\ArrayLoader(), 'en');
$validator_factory = new \Illuminate\Validation\Factory($translator, new \Illuminate\Container\Container());

$validator = $validator_factory->make(
    [
        'vendor' => 'yodlee',
        'advisor_id' => 1,
        'reference' => 'ABC123'
    ],
    [
        'vendor' => 'required|in:betterment,yodlee',
        'advisor_id' => 'required|numeric',
        'reference' => 'nullable|required_if:vendor,yodlee'
    ],
    [
        'vendor' => 'Yodlee is needed',
        'advisor_id' => 'Advisor ID is required and numeric',
        'reference' => 'Reference is required if vendor is yodlee',
    ],
    [
        'type' => 'nullable|required_if:vendor,betterment'
    ]
);

try {
    $input = $validator->validate();
} catch (\Illuminate\Validation\ValidationException $exception) {
    var_dump($validator->failed(), $exception->errors());die();
}

var_dump(1);



/**
 * Symfony Validation: Constraints, Validator
 */
/** @var \Symfony\Component\Validator\Validator\RecursiveValidator $symfony_validator */
$symfony_validator = \Symfony\Component\Validator\Validation::createValidator();
/** @var \Symfony\Component\Validator\ConstraintViolationList $violations */
$violations = $symfony_validator->validate('test', [
    new \Symfony\Component\Validator\Constraints\Length(['min' => 5]),
]);

if ($violations->count() > 0) {
    /** @var \Symfony\Component\Validator\ConstraintViolation $violation */
    foreach ($violations as $violation) {
        var_dump($violation->getMessage(), $violation->getCode());
    }
}