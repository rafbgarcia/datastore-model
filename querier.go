package db

import (
	"appengine"
	"reflect"
)

// Querier provides hight level query operations
type Querier struct {
	c appengine.Context
	q *Query
}

// All loads all matched items into the given slice
//
// It supports up to 1000 results sinces it delegates
// the job to datastore.Query.GetAll which imposes
// this limitation
//
// Future implementation will work with channels
// to overcome this limitation and let the users
// themselves dictate this limitation
func (this Querier) All(slice interface{}) error {
	keys, err := this.q.GetAll(this.c, slice)

	if err != nil {
		return err
	}

	for i, key := range keys {
		EntityAt(slice, i).SetKey(key)
	}

	return nil
}

// First loads the first matched item into the
// given entity
//
// datastore.Done is returned if there is no
// matched item
func (this Querier) First(entity entity) error {
	i := this.q.Run(this.c)
	key, err := i.Next(entity)
	if err != nil {
		return err
	}

	entity.SetKey(key)
	return nil
}

// EntityAt retrieves the entity type at position i
// in the given slice
//
// It panics in case the slice parameter is not either
// a slice or a slice pointer
func EntityAt(slice interface{}, i int) entity {
	s := reflect.ValueOf(slice)

	if s.Kind() == reflect.Slice {
		return s.Index(i).Interface().(entity)
	}

	if s.Kind() == reflect.Ptr && s.Type().Elem().Kind() == reflect.Slice {
		return s.Elem().Index(i).Interface().(entity)
	}

	panic("Querier.toEntitySlice given a non-slice type")
}