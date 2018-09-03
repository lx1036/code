<?php


namespace Next\Database\Orm;

use Illuminate\Contracts\Support\Arrayable;
use Illuminate\Support\Carbon;
use Illuminate\Support\Str;
use JsonSerializable;
use ArrayAccess;
use Next\Database\Orm\Builder;
use Next\Database\Orm\Query\Builder as QueryBuilder;
use Next\Database\Orm\Relations\BelongsTo;
use Next\Database\Orm\Relations\BelongsToMany;
use Next\Database\Orm\Relations\HasMany;
use Next\Database\Orm\Relations\HasOne;
use Next\Database\Orm\Relations\MapMany;
use Next\Database\Orm\Relations\MapOne;
use Next\Database\Orm\Relations\MapToMany;
use Next\Database\Orm\Relations\MapToOne;
use Next\Events\DispatcherInterface as Dispatcher;

abstract class Model implements JsonSerializable, ArrayAccess
{
    protected $attributes = [];

    protected $relations = [];

    protected static $booted = [];

    protected static $globalScopes = [];

    public $table;

    protected $fillable = [];

    protected $dates = [];

    protected $dateFormat;

    protected $visiable = [];
    protected $hidden = [];
    
    protected $connection;

    public function __construct(array $attributes = [])
    {
        // boot -> sync original -> fill attributes
        /**
         * boot is a place where can add a interceptor to query model when newly create a model
         * e.g. add a SoftDelete global scope, add a Readable global scope.
         */
        $this->bootIfNotBooted();

        $this->fill($attributes);
    }

    public function fill(array $attributes): Model
    {
        foreach ($this->getFillableFromArray($attributes) as $key => $value) {
            $key = $this->removeTableFromKey($key);

            if ($this->isFillable($key)) {
                $this->setAttribute($key, $value);
            }
        }

        return $this;
    }

    public function setAttribute(string $key, $value)
    {
        // set mutator
        if ($this->isSetMutator($key)) {
            return $this->setMutatorAttributeValue($key, $value);
        } elseif ($value && $this->isDateAttribute($key)) { // date
            $value = $this->formatDateValue($value);
        }

        $this->attributes[$key] = $value;

        return $this;
    }

    public function formatDateValue($value)
    {
        if (is_string($value)) {

        } elseif ($value instanceof Carbon) {
            $value = Carbon::create()->format($this->getDateFormat());
        }

        return $value;
    }

    public function getDateFormat()
    {
        return $this->dateFormat;
    }

    public function getDates()
    {
        return $this->dates;
    }

    public function isDateAttribute($key)
    {
        return in_array($key, $this->getDates(), true);
    }

    protected function isSetMutator($key)
    {
        return method_exists($this, 'set' . Str::studly($key) . 'Attribute');
    }


    public function setMutatorAttributeValue($key, $value): Model
    {
        return $this->{'set' . Str::studly($key) . 'Attribute'}($value);
    }

    public function isFillable(string $key): bool
    {
        return in_array($key, $this->getFillable(), true);
    }

    public function removeTableFromKey(string $key): string
    {
        return Str::contains($key, '.') ? last(explode('.', $key)) : $key;
    }

    public function getFillable(): array
    {
        return $this->fillable;
    }

    public function getFillableFromArray(array $attributes): array
    {
        if (count($this->getFillable()) > 0) {
            return array_intersect_key($attributes, array_flip($this->getFillable()));
        }

        return $attributes;
    }

    protected function bootIfNotBooted()
    {
        if (! isset(static::$booted[static::class])) {

            $this->fireModelEvent('booting');

            static::bootTraits();

            $this->fireModelEvent('booted');


            static::$booted[static::class] = true;
        }
    }

    protected static function bootTraits()
    {
        $class = static::class;

        foreach (class_uses_recursive($class) as $trait) {
            $method = 'boot' . class_basename($trait);

            if (method_exists($class, $method)) {
                forward_static_call([$class, $method]);
            }
        }
    }

    /**
     * @param \Closure|Scope $scope
     */
    public static function addGlobalScope($scope)
    {
        if ($scope instanceof Scope) {
            static::$globalScopes[static::class][get_class($scope)] = $scope;
        } elseif ($scope instanceof \Closure) {
            static::$globalScopes[static::class][spl_object_hash($scope)] = $scope;
        }

        throw new \InvalidArgumentException('Global scope must be instance of Scope or Closure');
    }

    public function getGlobalScopes()
    {
        return static::$globalScopes[static::class];
    }

    /**
     * @param Builder $builder
     * @return Builder
     */
    public function registerGlobalScopes(Builder $builder): Builder
    {
        foreach ($this->getGlobalScopes() as $identifier => $scope) {
            $builder->withGlobalScope($identifier, $scope);
        }

        return $builder;
    }

    /**
     * @return Builder
     */
    public function newQuery(): Builder
    {
        return $this->registerGlobalScopes($this->newQueryWithoutScopes());
    }
    

    public function newQueryBuilder()
    {
        return new QueryBuilder($connection = $this->getConnection(), $connection->getQueryGramma(), $connection->getQueryProcessor());
    }

    public function getConnection()
    {
        
    }

    public function newQueryWithoutScopes()
    {

    }

    public function newQueryWithoutRelations()
    {

    }

    public function qualifiedColumn(string $column)
    {
        return $this->getTable() . '.' . $column;
    }

    public function getTable()
    {
        if (! isset($this->table)) {
            $this->table = Str::snake(Str::plural(class_basename($this)));
        }

        return $this->table;
    }


    //region relations
    /**
     * Model Relations:
     * Person <=> IdentityCard: Person hasOne IdentityCard, IdentityCard belongsTo Person
     * Person <=> Car: Person hasMany Car, Car belongsTo Person
     * Article <=> Tag: Article hasMany Tag, Tag belongsToMany Article
     *
     * Article => Author, Video => Author, Music => Author:
     * (Article, Video, Music) mapOne Author, Author mapToOne (Article, Video, Music)
     * Article => Comment, Video => Comment, Music => Comment:
     * (Article, Video, Music) mapMany Comment, Comment mapToOne (Article, Video, Music)
     *
     */


    /**
     * @param string $child
     * @param $foreign_key
     * @param $local_key
     * @return HasOne
     */
    public function hasOne(string $child, $foreign_key, $local_key): HasOne
    {
        /** @var Model $related */
        $related = new $child;

        return new HasOne($related->newQuery(), $this, $related->getTable() . '.' . $foreign_key, $local_key);
    }

    public function hasMany(): HasMany
    {

    }

    public function belongsTo(string $parent, $owner_key, $local_key): BelongsTo
    {
        /** @var Model $related */
        $related = new $parent;

        return new BelongsTo($related->newQuery(), $this, $local_key, $owner_key);
    }


    public function belongsToMany(): BelongsToMany
    {

    }

    public function mapOne(): MapOne
    {

    }

    public function mapMany(): MapMany
    {

    }

    public function mapToOne(): MapToOne
    {

    }

    public function mapToMany(): MapToMany
    {

    }
    //endregion



    //region event

    /**
     * @var Dispatcher
     */
    protected static $dispatcher;

    public function setEventDispatcher(Dispatcher $dispatcher)
    {
        static::$dispatcher = $dispatcher;
    }

    public function getEventDispatcher(): Dispatcher
    {
        return static::$dispatcher;
    }


    public function fireModelEvent(string $event)
    {
        if (! isset(static::$dispatcher)) {
            return true;
        }

        return static::$dispatcher->fire($event . ':' . static::class, $this);
    }


    //endregion

    //region serialization
    public function attributesToArray()
    {
        // 1. handle Date attribute
        $attributes = $this->addDateAttributes($this->attributes);


        // 2. handle Mutate attribute
        $attributes = $this->addMutateAttributes($attributes);


        //3. handle Cast attribute
        $attributes = $this->addCastAttributes($attributes);

        $attributes = $this->addAppendAttributes($attributes);

        return $attributes;
    }

    public function relationsToArray(): array
    {
        $attributes = [];

        foreach ($this->getArrayableItems($this->relations) as $key => $value) {



            $attributes[$key] = $value;
        }

        return $attributes;
    }


    public function getVisible()
    {
        return $this->visiable;
    }

    public function getHidden()
    {
        return $this->hidden;
    }

    /**
     * Get can-output attributes from all attributes.
     *
     * @param array $values
     * @return array
     */
    protected function getArrayableItems(array $values): array
    {
        if (count($this->getVisible()) > 0) {
            $values = array_intersect_key($values, $this->getVisible());
        }

        if (count($this->getHidden())) {
            $values = array_diff_key($values, $this->getHidden());
        }

        return $values;
    }

    /**
     *
     *
     * @return array
     */
    public function toArray(): array
    {
        return array_merge($this->attributesToArray(), $this->relationsToArray());
    }

    public function toJson($options = 0)
    {
        $json = json_encode($this->jsonSerialize(), $options);

        if (JSON_ERROR_NONE !== json_last_error()) {
            throw new \InvalidArgumentException('Error json_encode [' . get_class($this) . '] with primary key [' . $this->getKey() . '].');
        }

        return $json;
    }

    public function jsonSerialize(): array
    {
        return $this->toArray();
    }

    public function __toString()
    {
        return $this->toJson();
    }

    //region mutate/cast




    //endregion


    //region hide/visible attributes



    //endregion


    //region append value to json



    //endregion

    //endregion


    public function offsetExists($offset)
    {
        // TODO: Implement offsetExists() method.
    }

    public function offsetGet($offset)
    {
        // TODO: Implement offsetGet() method.
    }

    public function offsetSet($offset, $value)
    {
        // TODO: Implement offsetSet() method.
    }

    public function offsetUnset($offset)
    {
        // TODO: Implement offsetUnset() method.
    }
}