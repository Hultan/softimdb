package builder

import (
	"log"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type Builder struct {
	builder *gtk.Builder
}

// NewBuilder creates a gtk.Builder and wraps it in a Builder struct
func NewBuilder(glade string) (*Builder, error) {
	// Create a new builder
	b, err := gtk.BuilderNewFromString(glade)
	if err != nil {
		return nil, err
	}
	return &Builder{builder: b}, nil
}

// GetObject gets a gtk object by name
func (b *Builder) GetObject(name string) glib.IObject {
	if b.builder == nil {
		log.Fatal("No builder set!")
	}
	obj, err := b.builder.GetObject(name)
	if err != nil {
		log.Fatal(err)
	}

	return obj
}
