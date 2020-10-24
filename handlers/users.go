package handlers 

import (
	"context"
	"net/http"

	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/go-playground/validator.v9"

	"github.com/inerts73/tronicscorp/dbiface"
	"github.com/labstack/echo"
)

//User represents a user
type User struct {
	Email		string `json:"username" bson:"username" validate:"required,email"`
	Password	string `json:"password" bson:"password" validate:"required,min=8,max=300"`
}

//UsersHandler users handler
type UsersHandler struct {
	Col dbiface.CollectionAPI
}

type userValidator struct {
	validator *validator.Validate
}

func (u *userValidator) Validate(i interface{}) error {
	return u.validator.Struct(i)
}

func insertUser(ctx context.Context, user User, collection dbiface.CollectionAPI) (interface{}, *echo.HTTPError) {
	var newUser User
	res := collection.FindOne(ctx, bson.M{"username": user.Email})
	err := res.Decode(&newUser)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Errorf("Unable to decode retrieved user: %v", err)
		return nil, echo.NewHTTPError(500, "Unable to decode retrieved user")
	}
	if newUser.Email != "" {
		log.Errorf("User by %s already exists", user.Email)
		return nil, echo.NewHTTPError(400, "User already exists")
	}
	insertRes, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Errorf("Unable to insert the user :%+v", err)
		return nil, echo.NewHTTPError(500, "Unable to create the user")
	}
	return insertRes.InsertedID, nil
}
 
//CreateUser create a user
func (h *UsersHandler) CreateUser(c echo.Context) error {
	var user User
	c.Echo().Validator = &userValidator{validator: v}
	if err := c.Bind(&user); err != nil {
		log.Errorf("Unable to bind to yser struct.")
		return echo.NewHTTPError(400, "Unable to parse the request payload.")
	}
	if err := c.Validate(user); err != nil {
		log.Errorf("Unable to validate the requested body.")
		return echo.NewHTTPError(400, "Unable to validate request payload.")
	}
	insertedUserID, err := insertUser(context.Background(), user, h.Col)
	if err != nil {
		log.Errorf("Unable to insert to database.")
		return err
	}
	return c.JSON(http.StatusCreated, insertedUserID)	
}