<?php


namespace Next\Database\Orm\Relations;


use Next\Database\Orm\Builder;

abstract class Relation
{
    protected $query;

    public function __construct(Builder $query)
    {
        $this->query = $query;
    }

    public function __call($method, $parameters)
    {
        return $this->query->{$method}(...$parameters);
    }
}