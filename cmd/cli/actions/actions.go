package actions

import (
	"context"
	"errors"
	"fmt"
)

// Key is a context key
type Key string

// FromContext returns key's value from given Context
func FromContext(ctx context.Context, key string) interface{} {
	return ctx.Value(Key(key))
}

// ToContext sets given value to Context's key
func ToContext(ctx context.Context, key string, value interface{}) context.Context {
	return context.WithValue(ctx, Key(key), value)
}

// PrintError print error details
func PrintError(err error) {
	fmt.Println("Error:", err)
}

// Action represents a cli action
type Action interface {
	Run(context.Context, []string) context.Context
	Help()
	Tips()
}

// Actions represents available cli actions
var Actions = map[string]Action{
	"backends": &Backends{},
	"backend":  &Backend{},
	"search":   &Search{},
	"manga":    &Manga{},
	"download": &Download{},
	"chapters": &Chapters{},
}

// Run execute cli action
func Run(ctx context.Context, action string, parameters []string) context.Context {
	a, ok := Actions[action]
	if !ok {
		PrintError(errors.New("action not recognized"))
		return ctx
	}

	ctx = a.Run(ctx, parameters)
	a.Tips()
	return ctx
}
