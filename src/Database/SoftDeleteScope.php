<?php


namespace Next\Database;


use Next\Database\Orm\Builder;

class SoftDeleteScope  implements Scope
{
    public function apply(Model $model, Builder $builder): void
    {
        $builder->whereNull($model->getQualifiedDeletedAtColumn());
    }


}