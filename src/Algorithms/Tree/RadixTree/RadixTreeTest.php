<?php


namespace Next\Algorithms\Tree\RadixTree;


use PHPUnit\Framework\TestCase;

class RadixTreeTest extends TestCase
{
    private $tree;

    protected function setUp(): void
    {
        parent::setUp();

        $this->tree = new Tree();
    }

    public function testAdd()
    {
        $this->tree->add('hello');
        self::assertEquals(1, $this->tree->wordCount());
        $this->tree->add('bye');
        self::assertEquals(8, $this->tree->size());
        self::assertEquals(2, $this->tree->wordCount());

        $this->tree->add('hello');
        self::assertEquals(8, $this->tree->size());
        self::assertEquals(2, $this->tree->wordCount());

        $this->tree->add('hell');
        self::assertEquals(8, $this->tree->size());
        self::assertEquals(3, $this->tree->wordCount());
    }

    public function testContains()
    {
        $this->tree->add('hello');
        self::assertTrue($this->tree->contains('hello'));
        $this->tree->add('bye');
        self::assertTrue($this->tree->contains('bye'));
        self::assertFalse($this->tree->contains('what'));
    }

    public function testStartsWith()
    {
        self::assertFalse($this->tree->startsWith('hello'));
        $this->tree->add('hello');
        $this->tree->add('bye');
        self::assertTrue($this->tree->startsWith('b'));
        self::assertTrue($this->tree->startsWith('he'));
        self::assertFalse($this->tree->startsWith('helloo'));
    }

    public function testWithPrefix()
    {
        $this->tree->add('hello');
        $this->tree->add('hell');
        $this->tree->add('bye');
        $this->tree->add('beyond');
        $withH = $this->tree->withPrefix('he');
        $withB = $this->tree->withPrefix('b');
        $withBy = $this->tree->withPrefix('by');
        $all = $this->tree->withPrefix('');
        self::assertSame(['hell', 'hello'], $withH);
        self::assertSame(['bye', 'beyond'], $withB);
        self::assertSame(['bye'], $withBy);
        self::assertSame(['hell', 'hello', 'bye', 'beyond'], $all);
        self::assertEquals(4, $this->tree->wordCount());
    }

    public function testDelete()
    {
        
    }

    public function testClear()
    {
        
    }

    public function testGetWords()
    {
        
    }
}
