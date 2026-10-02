package product

import (
	"errors"
	"foodeli/platform/middleware"
	"foodeli/platform/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, authMW gin.HandlerFunc) {
	products := r.Group("/products")

	products.GET("/:id", h.GetProductById)
	products.GET("/store/:storeId", h.ListProductsByStoreId)

	products.POST("", authMW, h.CreateProduct)
	products.PUT("/:id", authMW, h.UpdateProduct)
	products.DELETE("/:id", authMW, h.DeleteProduct)
}

// GetProductById godoc
// @Summary Get product by ID
// @Description Retrieve product information by UUID
// @Tags Products
// @Produce json
// @Param id path string true "Product UUID"
// @Success 200 {object} response.Response{data=Product}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /products/{id} [get]
func (h *Handler) GetProductById(c *gin.Context) {
	productID := c.Param("id")

	//is valid uuid
	id, err := uuid.Parse(productID)
	if err != nil {
		response.BadRequest(c, "Invalid product ID")
		return
	}

	//get product
	product, err := h.service.GetProductById(c, id)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			response.NotFound(c, "Product not found")
			return
		}
		response.InternalServerError(c, "Failed to get product: "+err.Error())
		return
	}

	response.Success(c, 200, product)

}

// ListProductsByStoreId godoc
// @Summary List products by store ID
// @Description Retrieve all products belonging to a store
// @Tags Products
// @Produce json
// @Param storeId path string true "Store UUID"
// @Success 200 {object} response.Response{data=[]Product}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /products/store/{storeId} [get]
func (h *Handler) ListProductsByStoreId(c *gin.Context) {
	storeID := c.Param("storeId")
	//is valid uuid
	id, err := uuid.Parse(storeID)
	if err != nil {
		response.BadRequest(c, "Invalid store ID")
		return
	}

	//get products
	products, err := h.service.ListProductsByStoreId(c, id)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			response.NotFound(c, "Store not found")
			return
		}
		response.InternalServerError(c, "Failed to get products: "+err.Error())
		return
	}

	response.Success(c, 200, products)
}

// CreateProduct godoc
// @Summary Create a product
// @Description Add a new product to merchant's store
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateProductRequest true "Product payload"
// @Success 201 {object} response.Response{data=Product}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /products [post]
func (h *Handler) CreateProduct(c *gin.Context) {
	merchantID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	//bind json
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	product, err := h.service.CreateProduct(c, merchantID, &req)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			response.NotFound(c, "Store not found")
			return
		}
		response.InternalServerError(c, "Failed to create product: "+err.Error())
		return
	}

	response.Success(c, 201, product)

}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update existing product details in merchant's store
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product UUID"
// @Param request body UpdateProductRequest true "Product update payload"
// @Success 200 {object} response.Response{data=Product}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /products/{id} [put]
func (h *Handler) UpdateProduct(c *gin.Context) {
	merchantId, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	productID := c.Param("id")
	//is valid uuid
	id, err := uuid.Parse(productID)
	if err != nil {
		response.BadRequest(c, "Invalid product ID")
		return
	}

	//bind json
	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	product, err := h.service.UpdateProduct(c, merchantId, id, &req)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			response.NotFound(c, "Product not found")
			return
		}
		if errors.Is(err, ErrUnauthorizedProductAccess) {
			response.Unauthorized(c, "You do not have permission to update this product")
			return
		}
		response.InternalServerError(c, "Failed to update product: "+err.Error())
		return
	}

	response.Success(c, 200, product)
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Soft-delete a product from merchant's store
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product UUID"
// @Success 200 {object} response.Response{data=string}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /products/{id} [delete]
func (h *Handler) DeleteProduct(c *gin.Context) {
	merchantID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	productID := c.Param("id")
	//is valid uuid
	id, err := uuid.Parse(productID)
	if err != nil {
		response.BadRequest(c, "Invalid product ID")
		return
	}

	err = h.service.DeleteProduct(c, merchantID, id)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			response.NotFound(c, "Product not found")
			return
		}
		if errors.Is(err, ErrUnauthorizedProductAccess) {
			response.Forbidden(c, "You do not have permission to delete this product")
			return
		}
		response.InternalServerError(c, "Failed to delete product: "+err.Error())
		return
	}

	response.Success(c, 200, "Product deleted successfully")
}
