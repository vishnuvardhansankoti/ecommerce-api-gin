package router

import (
	"database/sql"
	"net/http"

	"ecommerce-api-gin/internal/handlers"
	"ecommerce-api-gin/internal/services"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New(db *sql.DB) *gin.Engine {
	r := gin.Default()

	categoryHandler := handlers.NewCategoryHandler(services.NewCategoryService(db))
	productHandler := handlers.NewProductHandler(services.NewProductService(db))
	customerHandler := handlers.NewCustomerHandler(services.NewCustomerService(db))
	orderHandler := handlers.NewOrderHandler(services.NewOrderService(db))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	v1 := r.Group("/api/v1")
	{
		categories := v1.Group("/categories")
		{
			categories.GET("", categoryHandler.List)
			categories.POST("", categoryHandler.Create)
			categories.GET("/:id", categoryHandler.Get)
			categories.PUT("/:id", categoryHandler.Update)
			categories.DELETE("/:id", categoryHandler.Delete)
		}

		products := v1.Group("/products")
		{
			products.GET("", productHandler.List)
			products.POST("", productHandler.Create)
			products.GET("/:id", productHandler.Get)
			products.PUT("/:id", productHandler.Update)
			products.DELETE("/:id", productHandler.Delete)
		}

		customers := v1.Group("/customers")
		{
			customers.GET("", customerHandler.List)
			customers.POST("", customerHandler.Create)
			customers.GET("/:id", customerHandler.Get)
			customers.PUT("/:id", customerHandler.Update)
			customers.DELETE("/:id", customerHandler.Delete)
		}

		orders := v1.Group("/orders")
		{
			orders.GET("", orderHandler.List)
			orders.POST("", orderHandler.Create)
			orders.GET("/:id", orderHandler.Get)
			orders.PATCH("/:id/status", orderHandler.UpdateStatus)
		}
	}

	return r
}
