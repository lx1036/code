<?php


namespace Next\Database\Orm;

use Next\Database\Orm\Query\Builder as Query;

class Builder
{
    protected $query;

    protected $scopes = [];

    public function __construct(Query $query)
    {
        $this->query = $query;
    }

    public function __call($method, $parameters)
    {

        $this->query->{$method}(...$parameters);
    }

    public function withGlobalScope($identifier, $scope): Builder
    {
        $this->scopes[$identifier] = $scope;

        return $this;
    }

    /**
     * Important!!!
     *
     * @param array $columns
     */
    public function get($columns = ['*'])
    {

    }
}