package types

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return toCamelJSONName(fld.Name)
		}
		return name
	})
	_ = v.RegisterValidation("date", validateDateField)
}

func toCamelJSONName(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func validateDateField(fl validator.FieldLevel) bool {
	f := fl.Field()
	if f.Kind() == reflect.Ptr {
		if f.IsNil() {
			return true
		}
		f = f.Elem()
	}
	str := f.String()
	if str == "" {
		return true
	}
	_, err := time.Parse(time.DateOnly, str)
	return err == nil
}

func getJSONFieldName(f reflect.StructField) string {
	name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
	if name == "" || name == "-" {
		return toCamelJSONName(f.Name)
	}
	return name
}

// ListQuery parses page / pageSize / search from query string.
type ListQuery struct {
	Page     *int
	PageSize *int
	Search   *string
}

// FetchListQuery is a thin counterpart to FetchParams (no coupling to global config types).
func FetchListQuery(c *gin.Context) *ListQuery {
	out := &ListQuery{}
	if v, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		out.Page = &v
	}
	if v, err := strconv.Atoi(c.DefaultQuery("pageSize", "20")); err == nil {
		out.PageSize = &v
	}
	if s := strings.TrimSpace(c.DefaultQuery("search", "")); s != "" {
		out.Search = &s
	}
	return out
}

// FetchParams is an alias for FetchListQuery.
func FetchParams(c *gin.Context) *ListQuery {
	return FetchListQuery(c)
}

// Require reports ok; if false responds 400 with msg.
func Require(c *gin.Context, ok bool, msg string) bool {
	if !ok {
		BadRequest(c, &ErrOpts{Message: msg})
		return false
	}
	return true
}

// BindJSON parses JSON into ptr or writes 400 and returns false.
func BindJSON(c *gin.Context, ptr any) bool {
	if err := c.ShouldBindJSON(ptr); err != nil {
		BadRequest(c, &ErrOpts{Message: summarizeBind(err)})
		return false
	}
	return true
}

// Guard is an alias for BindJSON (legacy name).
func Guard(c *gin.Context, ptr any) bool {
	return BindJSON(c, ptr)
}

// BindCreate binds JSON and validates struct tags (`required:"Create"` on pointers or values).
func BindCreate[T any](c *gin.Context) *T {
	var obj T
	if !BindJSON(c, &obj) {
		return nil
	}
	if msg := checkCreateRequired(reflect.ValueOf(&obj)); msg != "" {
		BadRequest(c, &ErrOpts{Message: msg})
		return nil
	}
	return &obj
}

// BindUpdate binds JSON (PATCH-style). No Create-required enforcement.
func BindUpdate[T any](c *gin.Context) *T {
	var obj T
	if !BindJSON(c, &obj) {
		return nil
	}
	return &obj
}

// IsCreate / IsUpdate match legacy names.
func IsCreate[T any](c *gin.Context) *T { return BindCreate[T](c) }
func IsUpdate[T any](c *gin.Context) *T { return BindUpdate[T](c) }

func checkCreateRequired(pv reflect.Value) string {
	v := pv.Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get("required") != "Create" {
			continue
		}
		fv := v.Field(i)
		empty := false
		switch fv.Kind() {
		case reflect.Ptr:
			empty = fv.IsNil() || fv.Elem().IsZero()
		default:
			empty = fv.IsZero()
		}
		if empty {
			return getJSONFieldName(f) + " is required."
		}
	}
	return ""
}

func summarizeBind(err error) string {
	var ves validator.ValidationErrors
	if errors.As(err, &ves) {
		msgs := make([]string, 0, len(ves))
		for _, e := range ves {
			msgs = append(msgs, shortValidationErr(e.Field(), e.Tag()))
		}
		return strings.Join(msgs, "; ")
	}

	var syntax *json.SyntaxError
	if errors.As(err, &syntax) {
		return "Invalid JSON syntax."
	}

	var ut *json.UnmarshalTypeError
	if errors.As(err, &ut) {
		f := ut.Field
		if f == "" {
			f = extractFieldRegex(err.Error())
		}
		return f + " has invalid type."
	}

	s := err.Error()
	if strings.Contains(s, "unknown field") {
		return "Unknown JSON field."
	}
	return "Invalid request body."
}

func shortValidationErr(field, tag string) string {
	switch tag {
	case "required":
		return field + " is required."
	case "date":
		return field + " must be date (YYYY-MM-DD)."
	case "email":
		return field + " must be a valid email."
	case "uuid":
		return field + " must be a UUID."
	case "url", "uri":
		return field + " must be a valid URL."
	case "max", "min", "len", "gte", "lte", "gt", "lt", "eq", "oneof":
		return field + " failed validation (" + tag + ")."
	default:
		return field + " failed validation (" + tag + ")."
	}
}

var reStructFieldName = regexp.MustCompile(`struct field \w+\.(\w+) of type`)

func extractFieldRegex(s string) string {
	if m := reStructFieldName.FindStringSubmatch(s); len(m) > 1 {
		return toCamelJSONName(m[1])
	}
	return ""
}
