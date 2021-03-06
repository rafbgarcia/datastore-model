package db

import (
	"reflect"
	"strings"
	"appengine/datastore"
	"appengine"
)

// TODO might be better to have a Metadata type
// to encapsulate the key components and returned from
// a call to ExtractMetadataFrom(e) rather than
// holding state in the KeyResolver
//
// This way a single instance of KeyResolver might be
// used in parallel computations
type KeyResolver struct {
	context appengine.Context
	Kind      string
	StringID string
	IntID    int64
	HasParent bool
	Parent   *datastore.Key
}

// NewKeyResolver creates a new instance of *KeyResolver
func NewKeyResolver(c appengine.Context) *KeyResolver {
	return &KeyResolver{context: c}
}

// IsAutoGenerated tells whether or not a resolved key
// is auto generated by datastore
//
// Keys are auto generated if no struct field is tagged with db:"id"
func (this *KeyResolver) IsAutoGenerated() bool {
	return this.IntID == 0 && this.StringID == ""
}

// Resolve resolves the datastore key for the given entity
// by either assembling it based on the structs tags
// or by creating an auto generated key in case no tags are
// provided
//
// ErrMissingStringId is returned in case a string field
// is tagged with db:"id" and is empty
//
// ErrMissingIntId is returned in case an int field
// is tagged with db:"id" and is 0
func (this *KeyResolver) Resolve(e entity) error {
	if e.HasKey() {
		this.Parent = e.Key().Parent()
		this.HasParent = this.Parent != nil
		this.IntID = e.Key().IntID()
		this.StringID = e.Key().StringID()
		this.Kind = e.Key().Kind()
		return nil
	}

	if err := this.ExtractMetadataFrom(e); err != nil {
		return err
	}

	e.SetKey(datastore.NewKey(
		this.context,
		this.Kind,
		this.StringID,
		this.IntID,
		this.Parent,
	))

	return nil
}

func (this *KeyResolver) ExtractMetadataFrom(e entity) error {
	if err := this.ExtractKindMetadata(e); err != nil {
		return err
	}
	if err := this.ExtractKeyMetadata(e); err != nil {
		return err
	}
	return nil
}

// ExtractKeyMetadata extracts metadata from struct tags
// in order to resolve the datastore key for a given entity
//
// e.g.:
//
// The following struct declares an id tag on a field
// of type string, thus its StringID.
//
// type Person struct {
//   db.Model    `db:"People"`
//   Name string `db:"id"`
// }
//
// The following struct declares an id tag on a field
// of type int, thus its IntID.
//
// type BankAccount struct {
//   db.Model   `db:"Accounts"`
//   Number int `db:"id"`
// }
//
// If multiple id tags are used on a struct fields
// only the first tag from top to bottom is considered
func (this *KeyResolver) ExtractKeyMetadata(e entity) error {
	this.Parent = e.Parent()
	elem := reflect.TypeOf(e).Elem()
	elemValue := reflect.ValueOf(e).Elem()

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		tag := field.Tag.Get("db")
		value := elemValue.Field(i)
		if tag == "id" {
			switch field.Type.Kind() {
			case reflect.String:
				v := value.String()
				if v == "" {
					return ErrMissingStringId
				}
				this.StringID = v
				this.IntID = 0
				return nil
			case reflect.Int,
				reflect.Int8,
				reflect.Int16,
				reflect.Int32,
				reflect.Int64:
				v := value.Int()
				if v == 0 {
					return ErrMissingIntId
				}
				this.StringID = ""
				this.IntID = v
				return nil
			}
		}
	}

	// Default key values for auto generated keys
	return nil
}

// ExtractEntityKindMetadata extracts entity kind from struct tag
// applied to db.Model field
//
// e.g.:
//
// type Person struct {
//   db.Model `db:"People"`
//   Name     string
// }
//
// Returns the entity kind and whether the entity has parent key
//
// TODO merge the logic below with ExtractKeyMetadata
// There is no need to iterate over the struct fields twice
// All the metadata can be extracted in a single run
func (this *KeyResolver) ExtractKindMetadata(e entity) error {
	elem := reflect.TypeOf(e).Elem()
	this.Kind = elem.Name()

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if field.Type.Name() == reflect.TypeOf(Model{}).Name() {
			if kindMetadata := field.Tag.Get("db"); kindMetadata != "" {
				values := strings.Split(kindMetadata, ",")
				if strings.TrimSpace(values[0]) != "" {
					this.Kind = values[0]
				}
				if len(values) > 1 {
					this.HasParent = strings.TrimSpace(values[1]) == "hasparent"
					if this.HasParent && e.Parent() == nil {
						return ErrMissingParentKey
					}
				}
				return nil
			}
		}
	}

	return nil
}
