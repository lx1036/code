package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

type Person struct {
	Name string
	Age int
	Sex int
}

type People struct {
	Name string "Name"
	Age int "Age"
	Sex int "Sex"
}

type Household struct {
	Name string
	Age int
}

func (household *Household) GetName() string  {
	return household.Name
}

func (household *Household) GetAge() int {
	return household.Age
}

func main()  {
	household := new(Household)
	household.Name = "household1"
	household.Age = 12
	fmt.Println(household.GetName(), household.GetAge())

	person := new(Person)
	person.Name = "test1"
	person.Age = 10
	person.Sex = 1

	person1 := &Person{"test2", 11, 0}

	fmt.Println(person, person1, unsafe.Sizeof(person1.Name), unsafe.Sizeof(person1.Age), unsafe.Sizeof(person1.Sex), unsafe.Sizeof(person1))

	people := People{
		Name: "test3",
		Age:  0,
		Sex:  0,
	}
	for i := 0; i < 3; i++  {
		reflectTag(people, i)
	}
}

func reflectTag(people People, i int)  {
	peopleField := reflect.TypeOf(people).Field(i)

	fmt.Println(peopleField.Tag, peopleField.Name)
}
