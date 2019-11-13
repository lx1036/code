package skiplist

import "testing"

func TestNewConcurrentSkipList(test *testing.T) {
	type args struct {
		level int
	}
	fixtures := []struct{
		name string
		args args
	}{
		{name:"test1", args: args{level:-1}},
		{name: "test2", args: args{level:64}},
	}
	for _, fixture := range fixtures {
		test.Run(fixture.name, func(test *testing.T) {
			if got, err := NewConcurrentSkipList(fixture.args.level); err != nil {
				test.Errorf("Failed to create concurrent skiplist because of %v, got: %v", err, got)
			}
		})
	}
}

func TestConcurrentSkipListSearch(test *testing.T) {

}

func TestConcurrentSkipListInsert(test *testing.T) {
	level := 8
	skipList, _ := NewConcurrentSkipList(level)
	
	test.Run("level", func(test *testing.T) {
		if skipList.Level() != level {
			test.Errorf("wrong level, want %d got %d", level, skipList.Level())
		}
	})
	
	length := int32(0)
	test.Run("length", func(test *testing.T) {
		if skipList.Length() != length {
			test.Errorf("wrong level, want %d got %d", length, skipList.Length())
		}
	})
	
	for i := 1; i <= 10;i++  {
		skipList.Insert(uint64(i), i)
	}
	
}
