package template

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
)

// Templater interface is able to write a template to a writer, based on a Language
type Templater interface {
	Template(w io.Writer, lang Language) error
}

// Kept outside the function so it can be changed for testing. Path here breaks in testing
var templatepath = "./template/templates/%s"

func newTemplate(name, path string) (*template.Template, error) {
	return template.New(name).
		Funcs(funcMap).
		ParseFiles(fmt.Sprintf(templatepath, path))
}

var errNoType = errors.New("type not stored in package level type declaration")

// Header is the file header
func Header(w io.Writer, lang Language) error {
	tmpl, err := newTemplate("header.tmpl", fmt.Sprintf("%s/header.tmpl", lang))
	if err != nil {
		return err
	}
	return tmpl.Execute(w, nil)
}

// Raw is a template with raw input in it
func Raw(w io.Writer, raw string) error {
	tmpl, err := template.New("raw").Parse(raw)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, nil)
}

// PackageType is a package-level type. Any package type will
// be templated with a full type creation statement and possibly a comment
type PackageType struct {
	Name    string
	Comment string
	Type    Templater
	Tag     string
}

func (t *PackageType) Template(w io.Writer, lang Language) error {
	tmpl, err := newTemplate("declaration.tmpl", fmt.Sprintf("%s/declaration.tmpl", lang))
	if err != nil {
		return err
	}
	if err = tmpl.Execute(w, t); err != nil {
		return err
	}
	if t.Type == nil {
		log.WithError(errNoType).WithField("name", t.Name).Error("error while writing package type")
		return errNoType
	}

	return t.Type.Template(w, lang)
}

// Basic is a basic type. Ints, strings, bools, etc.. or a custom type.
type Basic struct {
	Type    string
	Pointer bool
}

func (t *Basic) Template(w io.Writer, lang Language) error {
	tmpl, err := newTemplate("basic.tmpl", fmt.Sprintf("%s/basic.tmpl", lang))
	if err != nil {
		return err
	}
	return tmpl.Execute(w, t)
}

type Map struct {
	Key   Templater
	Value Templater
}

func (t *Map) Template(w io.Writer, lang Language) error {
	tmpl, err := newTemplate("map_key.tmpl", fmt.Sprintf("%s/map_key.tmpl", lang))
	if err != nil {
		return err
	}
	if err = tmpl.Execute(w, t); err != nil {
		return err
	}
	if err := t.Key.Template(w, lang); err != nil {
		return err
	}
	tmpl, err = newTemplate("map_value.tmpl", fmt.Sprintf("%s/map_value.tmpl", lang))
	if err != nil {
		return err
	}
	if err = tmpl.Execute(w, t); err != nil {
		return err
	}

	if err = t.Value.Template(w, lang); err != nil {
		return err
	}
	tmpl, err = newTemplate("map_close.tmpl", fmt.Sprintf("%s/map_close.tmpl", lang))
	if err != nil {
		return err
	}
	return tmpl.Execute(w, t)
}

// Array has a type
type Array struct {
	Type Templater
}

func (t *Array) Template(w io.Writer, lang Language) error {
	tmpl, err := newTemplate("array_open.tmpl", fmt.Sprintf("%s/array_open.tmpl", lang))
	if err != nil {
		return err
	}
	if err = tmpl.Execute(w, t); err != nil {
		return err
	}
	if err = t.Type.Template(w, lang); err != nil {
		return err
	}
	tmpl, err = newTemplate("array_close.tmpl", fmt.Sprintf("%s/array_close.tmpl", lang))
	if err != nil {
		return err
	}
	return tmpl.Execute(w, t)
}

// Struct only has fields
type Struct struct {
	Fields []Field

	// Strict is just for Flow types.
	Strict bool

	// Embedded are the embedded types for a struct
	Embedded []string
}

func (t *Struct) Template(w io.Writer, lang Language) error {
	tmpl, err := newTemplate("struct_open.tmpl", fmt.Sprintf("%s/struct_open.tmpl", lang))
	if err != nil {
		return err
	}
	if err = tmpl.Execute(w, t); err != nil {
		return err
	}
	for i, v := range t.Fields {
		if err = v.Template(w, lang); err != nil {
			return err
		}
		if i < len(t.Fields)-1 {
			tmpl, err = newTemplate("field_close.tmpl", fmt.Sprintf("%s/field_close.tmpl", lang))
			if err != nil {
				return err
			}
			if err := tmpl.Execute(w, nil); err != nil {
				return err
			}
			tmpl, err = newTemplate("comment.tmpl", fmt.Sprintf("%s/comment.tmpl", lang))
			if err != nil {
				return err
			}
			if err := tmpl.Execute(w, v); err != nil {
				return err
			}
		} else {
			tmpl, err = newTemplate("comment.tmpl", fmt.Sprintf("%s/comment.tmpl", lang))
			if err != nil {
				return err
			}
			if err := tmpl.Execute(w, v); err != nil {
				return err
			}
			Raw(w, "\n")
		}
	}
	tmpl, err = newTemplate("struct_close.tmpl", fmt.Sprintf("%s/struct_close.tmpl", lang))
	if err != nil {
		return err
	}
	return tmpl.Execute(w, t)
}

// Field is a struct field
type Field struct {
	Name    string
	Type    Templater
	Comment string
	Tag     string
}

func (t *Field) Template(w io.Writer, lang Language) error {
	jsonName := strings.Split(getTag("json", t.Tag), ",")[0]
	if jsonName != "" {
		t.Name = jsonName
	}
	tmpl, err := newTemplate("field_name.tmpl", fmt.Sprintf("%s/field_name.tmpl", lang))
	if err != nil {
		return err
	}
	if err = tmpl.Execute(w, t); err != nil {
		return err
	}

	// If there is an override type on the struct field
	override := strings.Split(getTag("tw", t.Tag), ",")
	switch len(override) {
	case 2:
		ptr, err := strconv.ParseBool(string(override[1]))
		if err != nil {
			log.WithError(err).Errorf("error parsing bool for type %s", t.Name)
		}
		t.Type = &Basic{
			Type:    string(override[0]),
			Pointer: ptr,
		}
	case 1:
		if string(override[0]) != "" {
			t.Type = &Basic{
				Type:    string(override[0]),
				Pointer: false,
			}
		}
	}
	return t.Type.Template(w, lang)
}

func getTag(tag string, tags string) string {
	loc := strings.Index(tags, fmt.Sprintf("%s:\"", tag))
	if loc > -1 {
		bs := []byte(tags)
		bs = bs[loc+len(tag)+2:]
		loc = strings.Index(string(bs), "\"")
		if loc == -1 {
			return ""
		}
		return string(bs[:loc])
	}
	return ""
}
