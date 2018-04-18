package user

import (
	"context"
	"time"

	"github.com/hamakn/go_ddd_webapp/src/app/domain/user"
	appDatastore "github.com/hamakn/go_ddd_webapp/src/app/infrastructure/datastore"
	"github.com/hamakn/go_ddd_webapp/src/app/infrastructure/db"
	"github.com/hamakn/go_ddd_webapp/src/app/infrastructure/fixture"
	"google.golang.org/appengine/datastore"
)

const (
	kind = "User"
)

type repository struct {
}

// NewRepository returns user.Repository
func NewRepository() user.Repository {
	return &repository{}
}

func (r *repository) GetAll(ctx context.Context) ([]*user.User, error) {
	users := []*user.User{}
	q := datastore.NewQuery(kind)

	_, err := db.GetAll(ctx, q, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*user.User, error) {
	u := &user.User{ID: id}
	err := db.Get(ctx, u)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, user.ErrNoSuchEntity
		}
		return nil, err
	}

	return u, nil
}

func (r *repository) Create(ctx context.Context, u *user.User) error {
	err := u.Validate()
	if err != nil {
		return user.ErrValidationFailed
	}

	return appDatastore.RunInTransaction(ctx, func(tctx context.Context) error {
		// check email uniqueness
		if !canTakeUserEmail(tctx, u.Email) {
			return user.ErrEmailCannotTake
		}

		// check screen_name uniquness
		if !canTakeUserScreenName(tctx, u.ScreenName) {
			return user.ErrScreenNameCannotTake
		}

		err := db.Put(tctx, u)
		if err != nil {
			return err
		}

		// take email
		userEmail := newUserEmail(u)
		err = takeUserEmail(tctx, userEmail)
		if err != nil {
			return user.ErrEmailCannotTake
		}

		// take screen_name
		userScreenName := newUserScreenName(u)
		err = takeUserScreenName(tctx, userScreenName)
		if err != nil {
			return user.ErrScreenNameCannotTake
		}

		return nil
	},
		true, // XG
	)
}

func (r *repository) Update(ctx context.Context, u *user.User) error {
	err := u.Validate()
	if err != nil {
		return user.ErrValidationFailed
	}

	return appDatastore.RunInTransaction(ctx, func(tctx context.Context) error {
		oldUser, err := r.GetByID(tctx, u.ID)
		if err != nil {
			return err
		}

		// userEmail
		if oldUser.Email != u.Email {
			err := updateUserEmail(tctx, u, oldUser.Email)
			if err != nil {
				return err
			}
		}

		// userScreenName
		if oldUser.ScreenName != u.ScreenName {
			err := updateUserScreenName(tctx, u, oldUser.ScreenName)
			if err != nil {
				return err
			}
		}

		// TODO: rollback if error occurred
		u.UpdatedAt = time.Now()

		err = db.Put(tctx, u)
		if err != nil {
			return err
		}

		return nil
	},
		true, // XG
	)
}

func (r *repository) Delete(ctx context.Context, u *user.User) error {
	return appDatastore.RunInTransaction(ctx, func(tctx context.Context) error {
		// lock user
		txu, err := r.GetByID(tctx, u.ID)
		if err != nil {
			return err
		}

		// userEmail
		err = deleteUserEmail(tctx, u.Email, txu.ID)
		if err != nil {
			return err
		}

		// userScreenName
		err = deleteUserScreenName(tctx, u.ScreenName, txu.ID)
		if err != nil {
			return err
		}

		err = db.Delete(tctx, txu)
		if err != nil {
			return err
		}

		return nil
	},
		true, // XG
	)
}

func (r *repository) CreateFixture(ctx context.Context) ([]*user.User, error) {
	users := []*user.User{}

	err := fixture.Load("users", &users)
	if err != nil {
		return nil, err
	}

	// NOTE: run in out of txn by datastore (25 entities) limit
	for _, u := range users {
		r.Create(ctx, u)
	}

	return users, nil
}
