package models

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id, omitempty" json:"id" xml:"id"`
	Name     string             `bson:"name" json:"name" xml:"name"`
	Email    string             `bson:"email" json:"email" xml:"email"`
	UserName string             `bson:"user_name omitempty" json:"userName" xml:"userName"`
	Password string             `bson:"password omitempty" json:"password" xml:"password"`
}

func (user *User) Add() (User, error) {

	ctx := context.TODO()
	db := getUserCollection()

	user.ID = primitive.NewObjectID()

	_, err := db.InsertOne(ctx, user)
	if err != nil {
		return User{}, errors.New(fmt.Sprintf("Kullanıcı kayıt edilirken hata oluştu hata şu şekilde \"%v\"", err))
	}
	return *user, nil

}

func (user *User) FindAll() ([]User, error) {
	ctx := context.TODO()
	db := getUserCollection()

	cursor, err := db.Find(ctx, bson.D{})
	if err != nil {
		return []User{}, err
	}

	var users []User
	if err = cursor.All(ctx, &users); err != nil {
		return []User{}, err
	}
	return users, nil
}

func (user User) FindById(id primitive.ObjectID) (User, error) {

	ctx := context.TODO()
	db := getUserCollection()
	err := db.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
