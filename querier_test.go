package db_test

import (
	"github.com/drborges/datastore-model"
	"github.com/drborges/goexpect"
	"testing"
)

var (
	diego = NewPerson("Diego", "Brazil")
	munjal = NewPerson("Munjal", "USA")
	people = People{diego, munjal}
)

func NewPerson(name, country string) *Person {
	person := new(Person)
	person.Name = name
	person.Country = country
	return person
}

func TestQuerierEntityAtSlicePtr(t *testing.T) {
	entity := db.EntityAt(&people, 1)

	expect := goexpect.New(t)
	expect(entity).ToBe(munjal)
}

func TestQuerierEntityAtSlice(t *testing.T) {
	entity := db.EntityAt(people, 0)

	expect := goexpect.New(t)
	expect(entity).ToBe(diego)
}