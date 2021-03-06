package db_test

import (
	"testing"
	"github.com/drborges/datastore-model"
	"github.com/drborges/goexpect"
	"reflect"
	"appengine/aetest"
	"appengine/datastore"
)

func TestKindExtractorExtractsKindFromNonTaggedModel(t *testing.T) {
	type Tag struct {
		db.Model
		Name string
	}

	tag := &Tag{}
	meta := &db.Metadata{}
	fieldModel := reflect.TypeOf(tag).Elem().Field(0)

	err := db.KindExtractor{tag, meta}.Extract(fieldModel)

	expect := goexpect.New(t)
	expect(err).ToBe(nil)
	expect(meta.Kind).ToBe("Tag")
}

func TestKindExtractorExtractsKindFromTag(t *testing.T) {
	type Tag struct {
		db.Model   `db:"Tags"`
		Name string
	}

	tag := &Tag{}
	meta := &db.Metadata{}
	fieldModel := reflect.TypeOf(tag).Elem().Field(0)

	err := db.KindExtractor{tag, meta}.Extract(fieldModel)

	expect := goexpect.New(t)
	expect(err).ToBe(nil)
	expect(meta.Kind).ToBe("Tags")
}

func TestKindExtractorAccpetsModelEmbeddedField(t *testing.T) {
	type Tag struct {
		db.Model
		Name string
	}

	tag := &Tag{}
	meta := &db.Metadata{}
	fieldModel := reflect.TypeOf(tag).Elem().Field(0)

	accepts := db.KindExtractor{tag, meta}.Accept(fieldModel)

	expect := goexpect.New(t)
	expect(accepts).ToBe(true)
}

func TestKindExtractorDoesNotAccpetNonModelEmbeddedField(t *testing.T) {
	type Tag struct {
		db.Model
		Name string
	}

	tag := &Tag{}
	meta := &db.Metadata{}
	fieldModel := reflect.TypeOf(tag).Elem().Field(1)

	accepts := db.KindExtractor{tag, meta}.Accept(fieldModel)

	expect := goexpect.New(t)
	expect(accepts).ToBe(false)
}

func TestKeyExtractorResolvesKeyMetadataInfoForEntityWithKeyAlreadySet(t *testing.T) {
	c, _ := aetest.NewContext(nil)
	defer c.Close()

	type Tag struct {
		db.Model
		Name string `db:"id"`
	}

	tag := &Tag{Name: "golang"}
	db.NewKeyResolver(c).Resolve(tag)
	resolverForEntityWithKey := db.NewKeyResolver(c)
	err := resolverForEntityWithKey.Resolve(tag)

	expect := goexpect.New(t)
	expect(err).ToBe(nil)
	expect(resolverForEntityWithKey.IntID).ToBe(int64(0))
	expect(resolverForEntityWithKey.Kind).ToBe("Tag")
	expect(resolverForEntityWithKey.HasParent).ToBe(false)
	expect(resolverForEntityWithKey.StringID).ToBe("golang")
	expect(resolverForEntityWithKey.IsAutoGenerated()).ToBe(false)
	expect(resolverForEntityWithKey.Parent).ToBe((*datastore.Key)(nil))
}
