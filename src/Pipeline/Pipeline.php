<?php


namespace Next\Pipeline;

interface PipelineInterface {
    public function pipe(callable $stage): PipelineInterface;
    public function __invoke($payload);
    public function process($payload);
}

class Pipeline implements PipelineInterface
{
    private $stages = [];

    public function __construct(callable ...$stages)
    {
        $this->stages = $stages;
    }

    public function pipe(callable $stage): PipelineInterface
    {
        $this->stages[] = $stage;

        return $this;
    }

    public function process($payload)
    {
        foreach ($this->stages as $stage) {
            $payload = $stage($payload);
        }

        return $payload;
    }

    public function __invoke($payload)
    {
        return $this->process($payload);
    }
}

interface StageInterface {
    public function __invoke($payload);
}

class AStage implements StageInterface {
    public function __invoke($payload)
    {
        return $payload + 1;
    }
}

class BStage implements StageInterface {
    public function __invoke($payload)
    {
        return $payload * 2;
    }
}

$pipeline1 = (new Pipeline)->pipe(new AStage)->pipe(new BStage);
$pipeline2 = (new Pipeline)->pipe($pipeline1)->pipe(new BStage)->pipe(new AStage);
$response = $pipeline2->process(10);

var_dump($response);