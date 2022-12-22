package controllers

import (
	"context"
	database "ecommerce/database"
	models "ecommerce/models"
	tokens "ecommerce/tokens"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)

	}
	return string(bytes)

}
func VerifyPassword(userpassword string, givenpassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(userpassword), []byte(givenpassword))
	valid := true
	msg := ""
	if err != nil {
		msg = "password incorrect"
		valid = false
	}
	return valid, msg
}
func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		}
		validaterr := validate.Struct(user)
		if validaterr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"err": validaterr.Error()})
			return
		}
		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": "error while counting email"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"err": "This email already exists"})
			return
		}
		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			c.JSON(500, gin.H{"err": err})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"err": "phone number already exist"})
			return
		}
		password := HashPassword(*user.Password)
		user.Password = &password
		user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_ID = user.ID.Hex()
		token, refreshtoken, err := tokens.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, user.User_ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		}
		user.Token = &token
		user.Refresh_Token = &refreshtoken
		user.Usercart = make([]models.ProductUser, 0)
		user.Address_Details = make([]models.Address, 0)
		user.Order_Status = make([]models.Order, 0)
		_, inserterr := UserCollection.InsertOne(ctx, user)
		if inserterr != nil {
			c.JSON(500, gin.H{"err": inserterr})
			return
		}
		defer cancel()
		c.JSON(http.StatusCreated, gin.H{"msg": "successfully signed in!"})

	}

}
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		var founduser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		}
		err := UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "email", Value: user.Email}}).Decode(&founduser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": "error occured "})
			return
		}
		isverify, msg := VerifyPassword(*user.Password, *founduser.Password)
		if !isverify {
			c.JSON(http.StatusBadRequest, gin.H{"err": "your password is incorrect"})
			fmt.Println(msg)
			return
		}
		token, refreshtoken, err := tokens.TokenGenerator(*founduser.Email, *founduser.First_Name, *founduser.Last_Name, founduser.User_ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": "cannot generate new token"})
			return
		}
		tokens.UpdateAllToken(token, refreshtoken, founduser.User_ID)
		defer cancel()
		c.JSON(http.StatusOK, founduser)

	}

}
func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var product models.Product
		if err := c.BindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		}
		product.Product_ID = primitive.NewObjectID()
		_, err := ProductCollection.InsertOne(ctx, product)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": "not inserted into collection"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, "successfully added")
	}

}
func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var productlist []models.Product
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		cursor, err := ProductCollection.Find(ctx, bson.D{{}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Something went wrong")
			return
		}
		err = cursor.All(ctx, &productlist)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)
		if err := cursor.Err(); err != nil {
			log.Println(err)
			c.JSON(400, "invalid")
			return
		}
		defer cancel()
		c.JSON(200, productlist)

	}

}
func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var searchproduct []models.Product
		params := c.Query("name")
		if params == "" {
			log.Println("no query params")
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"err": "invalid"})
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		searchresult, err := ProductCollection.Find(ctx, bson.M{"product_name": bson.M{"$regex": params}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, "something went wrong")
			return
		}
		err = searchresult.All(ctx, &searchproduct)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, "Invalid")
			return
		}
		defer searchresult.Close(ctx)
		if err := searchresult.Err(); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, "invalid")
			return
		}
		c.JSON(http.StatusOK, searchproduct)
	}

}
