package controllers

import (
	"context"
	"ecommerce/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.Query("id")
		if userid == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"err": "user id is empty"})
			c.Abort()
			return
		}
		user_id, err := primitive.ObjectIDFromHex(userid)
		if err != nil {
			c.JSON(500, "Internal server error")
			return
		}
		var address models.Address
		address.Address_id = primitive.NewObjectID()
		if err := c.BindJSON(&address); err != nil {
			c.JSON(http.StatusNotAcceptable, err.Error())
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		match := bson.D{primitive.E{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: user_id}}}}
		unwind := bson.D{primitive.E{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$address_details"}}}}
		group := bson.D{primitive.E{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$address_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: 1}}}}}}
		cursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{match, unwind, group})
		if err != nil {
			c.JSON(500, "Internal server error")
			return
		}
		var info []bson.M
		if err := cursor.All(ctx, &info); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		var size int32
		for _, value := range info {
			count := value["total"]
			size = count.(int32)
		}
		if size < 2 {
			filter := bson.M{"_id": user_id}
			update := bson.D{primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "address_details", Value: address}}}}
			_, err := UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				c.Header("Content-Type", "application/json")
				c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
				return
			}

		} else {
			c.JSON(http.StatusBadRequest, gin.H{"err": "cannot have more than two addressses"})
			return
		}
		defer cancel()
		ctx.Done()
		c.JSON(200, "Successfully added address")
	}

}
func EditHomeAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.Query("id")
		if userid == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"err": "invalid search index"})
			c.Abort()
			return
		}
		user_id, err := primitive.ObjectIDFromHex(userid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "something went wrong")
			return
		}
		var editaddress models.Address
		if err := c.BindJSON(&editaddress); err != nil {
			c.JSON(http.StatusInternalServerError, "internal server error")
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		filter := bson.D{primitive.E{Key: "_id", Value: user_id}}
		update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "address_details.0.house", Value: editaddress.House},
			{Key: "address_details.0.street", Value: editaddress.Street},
			{Key: "address_details.0.city", Value: editaddress.City},
			{Key: "address_details.0.pincode", Value: editaddress.Pincode}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "internal server error")
			return
		}
		defer cancel()
		ctx.Done()
		c.JSON(200, "Successfully edited")

	}

}
func EditWorkAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.Query("id")
		if userid == "" {
			c.JSON(http.StatusNotFound, "invalid id")
			c.Abort()
			return
		}
		user_id, err := primitive.ObjectIDFromHex(userid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "something went wrong")
			return
		}
		var newaddress models.Address
		if err := c.BindJSON(&newaddress); err != nil {
			c.JSON(http.StatusInternalServerError, "something went wrong")
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		filter := bson.M{"_id": user_id}
		update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "address_details.1.house", Value: newaddress.House},
			{Key: "address_details.1.street", Value: newaddress.Street},
			{Key: "address_details.1.city", Value: newaddress.City},
			{Key: "address_details.1.pincode", Value: newaddress.Pincode}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "something went wrong")
			return
		}
		defer cancel()
		ctx.Done()
		c.JSON(200, "Successfully edited workaddress")
	}

}
func DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.Query("id")
		if userid == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"err": "invalid search index"})
			c.Abort()
			return
		}
		address := make([]models.Address, 0)
		user_id, err := primitive.ObjectIDFromHex(userid)
		if err != nil {
			c.JSON(500, "Internal server error")
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		filter := bson.D{primitive.E{Key: "_id", Value: user_id}}
		update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "address_details", Value: address}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "error occurred while deleting")
			return
		}
		defer cancel()
		ctx.Done()
		c.JSON(http.StatusOK, "successfully deleted")
	}

}
