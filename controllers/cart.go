package controllers

import (
	database "ecommerce/database"
	"ecommerce/models"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"context"

	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")

type Application struct {
	productCollection *mongo.Collection
	userCollection    *mongo.Collection
}

func NewApplication(productCollection, userCollection *mongo.Collection) *Application {
	return &Application{
		productCollection: productCollection,
		userCollection:    userCollection,
	}
}

func (app *Application) AddToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		product_id := c.Query("id")
		if product_id == "" {
			c.Header("Content-Type", "application/json")
			log.Println("product id is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}
		user_id := c.Query("userid")
		if user_id == "" {
			c.Header("Content-Type", "application/json")
			log.Println("userid is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}
		productid, err := primitive.ObjectIDFromHex(product_id)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = database.AddProductToCart(ctx, app.productCollection, app.userCollection, productid, user_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, "successfully added to cart")

	}
}
func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		product_id := c.Query("id")
		if product_id == "" {
			c.Header("Content-Type", "application/json")
			log.Println("product id is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}
		user_id := c.Query("userid")
		if user_id == "" {
			c.Header("Content-Type", "application/json")
			log.Println("userid is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}
		productid, err := primitive.ObjectIDFromHex(product_id)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = database.RemoveCartItem(ctx, app.productCollection, app.userCollection, productid, user_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, "successfully removed item from cart")

	}

}
func (app *Application) GetItemFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.Query("id")
		if userid == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"err": "invalid id "})
			c.Abort()
			return
		}
		user_id, _ := primitive.ObjectIDFromHex(userid)
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var filledcart models.User

		err := UserCollection.FindOne(ctx, bson.M{"_id": user_id}).Decode(&filledcart)
		defer cancel()
		if err != nil {
			log.Println(err)
			c.JSON(500, "not found")
			return
		}
		filter_match := bson.D{primitive.E{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: user_id}}}}
		unwind := bson.D{primitive.E{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$user_cart"}}}}
		group := bson.D{primitive.E{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"},
			{Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$user_cart.price"}}}}}}

		cursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filter_match, unwind, group})
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		var list []bson.M
		if err := cursor.All(ctx, &list); err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		for _, value := range list {
			c.JSON(200, value["total"])
			c.JSON(200, filledcart.Usercart)
		}
		ctx.Done()
		defer cancel()
	}
}
func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.Query("id")
		if userid == "" {
			c.Header("Content-Type", "application/json")
			log.Println("user id is empty")
			c.AbortWithError(http.StatusInternalServerError, errors.New("user id is empty"))
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		err := database.BuyItemFromCart(ctx, app.userCollection, userid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(200, "Successfully you have buyed")
	}

}
func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(c *gin.Context) {
		product_id := c.Query("id")
		if product_id == "" {
			c.Header("Content-Type", "application/json")
			log.Println("product id is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}
		user_id := c.Query("userid")
		if user_id == "" {
			c.Header("Content-Type", "application/json")
			log.Println("userid is empty")
			c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}
		productid, err := primitive.ObjectIDFromHex(product_id)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = database.InstantBuyer(ctx, app.productCollection, app.userCollection, productid, user_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, "successfully placed the order")

	}
}
