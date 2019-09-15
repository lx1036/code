<?php


namespace Next\Database\Orm\Query;


use Next\Database\Orm\Query\Connection\ConnectionInterface;
use Next\Database\Orm\Query\Grammar\GrammarInterface;
use Next\Database\Orm\Query\PostProcessor\ProcessorInterface;

class Builder
{
    protected $wheres = [];

    /** @var ConnectionInterface  */
    protected $connection;

    /** @var GrammarInterface  */
    protected $grammar;

    /** @var ProcessorInterface  */
    protected $processor;

    protected $table;

    public function __construct(ConnectionInterface $connection, GrammarInterface $grammar, ProcessorInterface $processor)
    {
        $this->connection = $connection;
        $this->grammar = $grammar;
        $this->processor = $processor;
    }


    public function whereNull($column): Builder
    {
        $type = 'Null';

        $this->wheres[] = compact('column', 'type');

        return $this;
    }

    public function from(string $table)
    {
        $this->table = $table;

        return $this;
    }
}