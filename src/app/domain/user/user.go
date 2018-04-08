package user

import "time"

// User Entity
type User struct {
	ID         int64     `json:"id" yaml:"id" datastore:"-" goon:"id"`
	Email      string    `json:"email" yaml:"email" validate:"required,email"`
	ScreenName string    `json:"screen_name" yaml:"screen_name" validate:"required"`
	Age        int       `json:"age" yaml:"age" validate:"required,min=0,max=120"`
	CreatedAt  time.Time `json:"created_at" yaml:"created_at" validate:"required"`
	UpdatedAt  time.Time `json:"updated_at" yaml:"updated_at" validate:"required"`
}
