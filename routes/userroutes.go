package routes

import (
	"ecommerce/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingroutes *gin.Engine) {
	incomingroutes.POST("/users/signup", controllers.Signup())
	incomingroutes.POST("/users/login", controllers.Login())
	incomingroutes.POST("/admin/addproduct", controllers.ProductViewerAdmin())
	incomingroutes.GET("/users/productview", controllers.SearchProduct())
	incomingroutes.GET("/users/search", controllers.SearchProductByQuery())
}
