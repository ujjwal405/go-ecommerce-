package database

import (
	"context"
	"ecommerce/models"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCantFindProduct    = errors.New("cant't find the product")
	ErrCantDecodeProduct  = errors.New("can't decode the product")
	ErrUserIdIsNotValid   = errors.New("user is not valid")
	ErrCantUpdateUser     = errors.New("cannot update this product to cart")
	ErrCantRemoveItemCart = errors.New("cannot remove this item form cart")
	ErrCantGetItem        = errors.New("unable to get item from cart")
	ErrCantBuyCartItem    = errors.New("cannot update the purchase")
)

func AddProductToCart(ctx context.Context, prodcollection, usercollection *mongo.Collection, productid primitive.ObjectID, userid string) error {
	product, err := prodcollection.Find(ctx, bson.M{"_id": productid})
	if err != nil {
		return ErrCantFindProduct
	}
	var productcart []models.ProductUser
	err = product.All(ctx, &productcart)
	if err != nil {
		return ErrCantDecodeProduct
	}
	user_id, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return ErrUserIdIsNotValid
	}
	filter := bson.D{primitive.E{Key: "_id", Value: user_id}}
	update := bson.D{primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "user_cart", Value: bson.D{primitive.E{Key: "$each", Value: productcart}}}}}}
	_, err = usercollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return ErrCantDecodeProduct
	}
	return nil
}
func RemoveCartItem(ctx context.Context, usercollection, prodcollection *mongo.Collection, productid primitive.ObjectID, userid string) error {
	user_id, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return ErrUserIdIsNotValid
	}
	filter := bson.M{"_id": user_id}
	update := bson.M{"$pull": bson.M{"user_cart": bson.M{"product_id": productid}}}
	_, err = usercollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return ErrCantDecodeProduct
	}
	return nil
}
func BuyItemFromCart(ctx context.Context, usercollection *mongo.Collection, userid string) error {
	id, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}
	var getitemcart models.User
	var ordercart models.Order

	ordercart.Order_ID = primitive.NewObjectID()
	ordercart.Ordered_At = time.Now()
	ordercart.Order_Cart = make([]models.ProductUser, 0)
	ordercart.Payment_Method.COD = true
	unwind := bson.D{primitive.E{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$user_cart"}}}}
	group := bson.D{primitive.E{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$user_cart.price"}}}}}}
	result, err := usercollection.Aggregate(ctx, mongo.Pipeline{unwind, group})
	ctx.Done()
	if err != nil {
		panic(err)
	}
	var getusercart []bson.M
	if err = result.All(ctx, &getusercart); err != nil {
		panic(err)
	}
	var total int32
	for _, item := range getusercart {
		Price := item["total"]
		total = Price.(int32)
	}
	ordercart.Price = int(total)
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "order_status", Value: ordercart}}}}
	_, err = usercollection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println(err)
	}
	err = usercollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&getitemcart)
	if err != nil {
		log.Println(err)
	}
	filter1 := bson.D{primitive.E{Key: "_id", Value: id}}
	update1 := bson.M{"$push": bson.M{"order_status.$[].order_cart": bson.M{"$each": getitemcart.Usercart}}}
	_, err = usercollection.UpdateOne(ctx, filter1, update1)
	if err != nil {
		log.Println(err)
	}
	empty_cart := make([]models.ProductUser, 0)
	filter3 := bson.D{primitive.E{Key: "_id", Value: id}}
	update3 := bson.D{primitive.E{Key: "user_cart", Value: empty_cart}}
	_, err = usercollection.UpdateOne(ctx, filter3, update3)
	if err != nil {
		return ErrCantBuyCartItem
	}
	return nil
}
func InstantBuyer(ctx context.Context, prodcollection, usercollection *mongo.Collection, productid primitive.ObjectID, userid string) error {
	id, err := primitive.ObjectIDFromHex(userid)
	if err != nil {
		return ErrUserIdIsNotValid
	}
	var productdetails models.ProductUser
	var orderdetails models.Order
	orderdetails.Order_ID = primitive.NewObjectID()
	orderdetails.Ordered_At = time.Now()
	orderdetails.Order_Cart = make([]models.ProductUser, 0)
	orderdetails.Payment_Method.COD = true
	err = prodcollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: productid}}).Decode(&productdetails)
	if err != nil {
		log.Println(err)
	}
	orderdetails.Price = productdetails.Price
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "order_status", Value: orderdetails}}}}
	_, err = usercollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
	}
	filter1 := bson.D{primitive.E{Key: "_id", Value: id}}
	update1 := bson.M{"$push": bson.M{"order_status.$[].order_cart": productdetails}}
	_, err = usercollection.UpdateOne(ctx, filter1, update1)
	if err != nil {
		log.Println(err)
	}
	return nil
}
