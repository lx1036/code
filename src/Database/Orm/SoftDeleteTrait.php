<?php


namespace Next\Database\Orm;


trait SoftDeleteTrait
{
    public static function bootSoftDeleteTrait()
    {
        static::addGlobalScope(new SoftDeleteScope);
    }

    protected function getQualifiedDeletedAtColumn()
    {
        $this->qualifiedColumn($this->getDeletedAtColumn());
    }

    public function getDeletedAtColumn()
    {
        return defined('static::DELETED_AT') ? static::DELETED_AT : 'deleted_at';
    }
}