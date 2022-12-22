package tokens

import (
	"context"
	"ecommerce/database"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")

type SignedDetails struct {
	Email      string
	First_Name string
	Last_Name  string
	Uid        string
	jwt.StandardClaims
}

func TokenGenerator(email string, firstname string, lastname string, uid string) (Signedtoken, SignedRefreshtoken string, err error) {
	err = godotenv.Load(".env")
	if err != nil {
		return
	}
	var SECRET_KEY = os.Getenv("SECRET_KEY")
	claims := &SignedDetails{
		Email:      email,
		First_Name: firstname,
		Last_Name:  lastname,
		Uid:        uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}
	refreshclaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}
	Signedtoken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return
	}
	SignedRefreshtoken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshclaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return
	}
	return
}
func ValidateToken(signedtoken string) (claims *SignedDetails, msg string) {
	err := godotenv.Load(".env")
	if err != nil {
		return
	}
	var SECRET_KEY = os.Getenv("SECRET_KEY")

	token, err := jwt.ParseWithClaims(signedtoken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return SECRET_KEY, nil
	})
	if err != nil {
		msg = err.Error()
		return
	}
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "token is not valid"
		return
	}
	if claims.StandardClaims.ExpiresAt < time.Now().Local().Unix() {
		msg = "token already expires"
		return
	}
	return
}
func UpdateAllToken(signedtoken string, refreshtoken string, userid string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	filter := bson.M{"user_id": userid}
	updatedat, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	upsert := true
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	data := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "token", Value: signedtoken},
		{Key: "refresh_token", Value: refreshtoken},
		{Key: "updated_at", Value: updatedat}}}}
	_, err := UserCollection.UpdateOne(ctx, filter, data, &opt)
	defer cancel()
	if err != nil {
		log.Panic(err)
		return
	}
}
