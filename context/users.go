package context

import (
	"context"

	"github.com/etaseq/lenslocked/models"
)

// This is the type that I am going to use for my keys.
// It is a best practice to define custom types for keys
// to prevent accidental key name collisions between
// different parts of the application.
type key string

const (
	userKey key = "user"
)

// Store a User inside context
func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// Retrieve a User from the context
func User(ctx context.Context) *models.User {
	val := ctx.Value(userKey)

	// After I get the value I need to perform type assertion
	// The reason I have to do that is because the Value method
	// returns an empty interface{}.
	// ##### Value(key any) any ######
	user, ok := val.(*models.User)
	if !ok {
		// The most likely case is that nothing was ever stored in the context,
		// so it doesn't have a type of *models.User. It is also possible that
		// other code in this package wrote an invalid value using the user key.
		return nil
	}

	return user
}
