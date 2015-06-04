package db

import (
	"appengine"
	"appengine/datastore"
	"errors"
	"reflect"
)

var (
	// ErrNoSuchEntity returned whenever a non existent
	// key is used to retrieve, update or delete
	ErrNoSuchEntity = errors.New("Entity not found")

	// ErrEntityExists returned when creating an entity
	// with a key that is already being used
	ErrEntityExists = errors.New("Entity already exists")
)

type entity interface {
	HasKey() bool
	Key() *datastore.Key
	SetKey(*datastore.Key)
	Parent() *datastore.Key
	SetParent(*datastore.Key)
	UUID() string
	SetUUID(uuid string) error
}

// Datastore Service that provides a set of
// operations to make it easy on you when
// working with appengine datastore
//
// It works along with db.Model in order to
// provide its features.
type Datastore struct {
	Context appengine.Context
}

// Create creates a new entity in datastore
// using the key generated by the keyProvider
//
// ErrEntityExists is returned in case the key
// generated by the KeyProvider is already being
// used
func (this Datastore) Create(e entity) error {
	if err := this.Load(e); err == nil {
		return ErrEntityExists
	}

	key, err := datastore.Put(this.Context, this.NewKeyFor(e), e)
	e.SetKey(key)
	return err
}

// Update updated an entity in datastore
//
// ErrNoSuchEntity is returned if the given
// entity does not exist in datastore
func (this Datastore) Update(e entity) error {
	if err := this.Load(e); err != nil {
		return err
	}
	_, err := datastore.Put(this.Context, e.Key(), e)
	return err
}

// Load loads entity data from datastore
//
// In case the entity has no key yet assigned
// a new one is created by the entity itself
// and used to retrieve the entity data from
// datastore
//
// ErrNoSuchEntity is returned in case no
// entity is found for the given key
func (this Datastore) Load(e entity) error {
	if !e.HasKey() {
		e.SetKey(this.NewKeyFor(e))
	}
	err := datastore.Get(this.Context, e.Key(), e)
	if err == datastore.ErrNoSuchEntity {
		return ErrNoSuchEntity
	}
	return err
}

// Delete deletes an entity from datastore
//
// ErrNoSuchEntity is returned in case the
// key provided does not match any existent
// entity
func (this Datastore) Delete(e entity) error {
	if err := this.Load(e); err != nil {
		return err
	}

	if !e.HasKey() {
		e.SetKey(this.NewKeyFor(e))
	}

	return datastore.Delete(this.Context, e.Key())
}

// NewKeyFor generates a new datastore key for the given entity
//
// The Key components are derived from the entity struct through reflection
// Fields tagged with `db:"id"` are used in the key as a StringID if
// the field type is string, or IntID in case its type is any int type
//
// In case multiple fields are tagged with `db:"id"`, the first field
// is selected to be used as id in the key
//
// If no field is tagged, the key is generated using the default values
// for StringID and IntID, causing the key to be auto generated
func (this Datastore) NewKeyFor(e entity) *datastore.Key {
	kind := reflect.TypeOf(e).Elem().Name()
	stringID, intID := this.extractIDs(e)
	parentKey := e.Parent()
	return datastore.NewKey(this.Context, kind, stringID, intID, parentKey)
}

func (this Datastore) extractIDs(e entity) (string, int64) {
	elem := reflect.TypeOf(e).Elem()
	elemValue := reflect.ValueOf(e).Elem()

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		tag := field.Tag.Get("db")
		value := elemValue.Field(i)
		if tag == "id" {
			switch field.Type.Kind() {
			case reflect.String:
				return value.String(), 0
			case reflect.Int,
				reflect.Int8,
				reflect.Int16,
				reflect.Int32,
				reflect.Int64:
				return "", value.Int()
			}
		}
	}

	// Default key values for auto generated keys
	return "", 0
}