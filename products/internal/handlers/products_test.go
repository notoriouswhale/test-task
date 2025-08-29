package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"products/internal/apperrors"
	"products/internal/models"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) Create(ctx context.Context, product *models.CreateProductDTO) (*models.Product, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}
func (m *MockProductService) Delete(ctx context.Context, id string) (*models.Product, error) {
	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}
func (m *MockProductService) List(ctx context.Context, listDTO *models.ListProductsDTO) ([]models.Product, int, error) {
	args := m.Called(ctx, listDTO)
	if args.Get(0) == nil {
		return nil, 0, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Int(1), args.Error(2)
}

func setupTestHandler() (*MockProductService, *ProductsHandler) {
	mockService := &MockProductService{}
	handler := NewProductsHandler(mockService, zap.NewNop())
	return mockService, handler
}

func TestProductHandler_CreateProduct(t *testing.T) {
	type testCase struct {
		name           string
		body           string
		expectedStatus int
	}

	cases := []testCase{
		{
			name:           "Create Product - Success Without Description",
			body:           `{"name":"Test Product","price":100}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Create Product - Success With Description",
			body:           `{"name":"Test Product","price":100,"description":"Test Description"}`,
			expectedStatus: http.StatusCreated,
		},
	}
	// Arrange
	mockService, handler := setupTestHandler()

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			productReq := tCase.body
			req := httptest.NewRequest("POST", "/products", strings.NewReader(productReq))
			w := httptest.NewRecorder()

			// Mock service expectation
			product := &models.Product{ID: "8f293f9f-9bd0-4294-bd17-4fb80aa2650a", Name: "Test Product", Description: "Test Description", Price: 100}
			mockService.On("Create", mock.Anything, mock.Anything).Return(product, nil).Once()

			// Act
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			handler.Create(ctx)

			// Assert
			assert.Equal(t, http.StatusCreated, w.Code, "Expected status code 201 for successful product creation")
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Response should have JSON content type")
			mockService.AssertExpectations(t)

			// Unmarshal and assert response fields
			var resp struct {
				Success bool            `json:"success"`
				Data    *models.Product `json:"data"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err, "Response body should be valid JSON")
			assert.True(t, resp.Success, "Response should indicate success")
			assert.Equal(t, product, resp.Data, "Response data should match created product")
		})

	}
}
func TestProductHandler_CreateProduct_BadRequest(t *testing.T) {
	type testCase struct {
		name string
		body string
	}

	cases := []testCase{
		{
			name: "Create Product - Failure Short Name < 3 Characters",
			body: `{"name":"Te","price":100}`,
		},
		{
			name: "Create Product - Failure Long Name > 50 Characters",
			body: `{"name":"Test Product with a very long name that exceeds fifty characters","price":100}`,
		},
		{
			name: "Create Product - Failure Price <= 0",
			body: `{"name":"Test Product","price":0}`,
		},
		{
			name: "Create Product - Failure No Name",
			body: `{"price":100}`,
		},
	}
	// Arrange
	mockService, handler := setupTestHandler()

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {

			productReq := tCase.body
			req := httptest.NewRequest("POST", "/products", strings.NewReader(productReq))
			w := httptest.NewRecorder()

			// Act
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			handler.Create(ctx)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for bad request")
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Response should have JSON content type")
			mockService.AssertNotCalled(t, "Create")

			var resp struct {
				Success bool   `json:"success"`
				Error   string `json:"error"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err, "Response body should be valid JSON")
			assert.False(t, resp.Success, "Response should indicate failure")
			assert.NotEmpty(t, resp.Error, "Error message should not be empty")
		})

	}
}

func TestProductHandler_CreateProduct_InternalServerError(t *testing.T) {
	// Arrange
	mockService, handler := setupTestHandler()

	productReq := `{"name":"Test Product","price":100}`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(productReq))
	w := httptest.NewRecorder()

	mockService.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("some internal server error")).Once()
	// Act
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	handler.Create(ctx)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected status code 500 for internal server error")
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Response should have JSON content type")
	mockService.AssertExpectations(t)

	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "Response body should be valid JSON")
	assert.False(t, resp.Success, "Response should indicate failure")
	assert.Equal(t, "Internal Server Error", resp.Error, "Error message should be Internal Server Error")
}

func TestProductHandler_DeleteProduct(t *testing.T) {
	type testCase struct {
		name      string
		productID string
	}

	cases := []testCase{
		{
			name:      "Delete Product - Success",
			productID: "8f293f9f-9bd0-4294-bd17-4fb80aa2650a",
		},
	}
	// Arrange
	mockService, handler := setupTestHandler()

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/products/"+tCase.productID, nil)
			w := httptest.NewRecorder()

			// Mock service expectation
			product := &models.Product{ID: tCase.productID, Name: "Test Product", Description: "Test Description", Price: 100}
			mockService.On("Delete", mock.Anything, tCase.productID).Return(product, nil).Once()

			// Act
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Params = gin.Params{gin.Param{Key: "id", Value: tCase.productID}}
			handler.Delete(ctx)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 for successful deletion")
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Response should have JSON content type")
			mockService.AssertExpectations(t)

			// Unmarshal and assert response fields
			var resp struct {
				Success bool            `json:"success"`
				Data    *models.Product `json:"data"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err, "Response body should be valid JSON")
			assert.True(t, resp.Success, "Response should indicate success")
			assert.Equal(t, product, resp.Data, "Response data should match deleted product")
		})
	}
}
func TestProductHandler_DeleteProduct_NotFound(t *testing.T) {
	productID := "8f293f9f-9bd0-4294-bd17-4fb80aa2650a"
	// Arrange
	mockService, handler := setupTestHandler()

	req := httptest.NewRequest("DELETE", "/products/"+productID, nil)
	w := httptest.NewRecorder()

	// Mock service expectation
	var eErr error = &apperrors.ErrorNotFound{ID: productID}
	mockService.On("Delete", mock.Anything, productID).Return(nil, eErr).Once()

	// Act
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{gin.Param{Key: "id", Value: productID}}
	handler.Delete(ctx)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404 for not found")
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Response should have JSON content type")
	mockService.AssertExpectations(t)

	// Unmarshal and assert response fields
	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "Response body should be valid JSON")
	assert.False(t, resp.Success, "Response should indicate failure")
	assert.Equal(t, eErr.Error(), resp.Error, "Error message should match with not found error")
}

func TestProductHandler_DeleteProduct_BadRequest(t *testing.T) {
	invalidId := "8f293f9f-9bd0-4294-bd17-4fb80"
	// Arrange
	mockService, handler := setupTestHandler()

	req := httptest.NewRequest("DELETE", "/products/"+invalidId, nil)
	w := httptest.NewRecorder()

	// Act
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{gin.Param{Key: "id", Value: invalidId}}
	handler.Delete(ctx)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for bad request")
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Response should have JSON content type")
	mockService.AssertNotCalled(t, "Delete")

	// Unmarshal and assert response fields
	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "Response body should be valid JSON")
	assert.False(t, resp.Success, "Response should indicate failure")
	assert.NotEmpty(t, resp.Error, "Error message should not be empty")
}

func TestProductHandler_DeleteProduct_InternalError(t *testing.T) {
	productID := "8f293f9f-9bd0-4294-bd17-4fb80aa2650a"
	// Arrange
	mockService := &MockProductService{}
	handler := NewProductsHandler(mockService, zap.NewNop())

	req := httptest.NewRequest("DELETE", "/products/"+productID, nil)
	w := httptest.NewRecorder()

	// Mock service expectation
	var eErr error = errors.New("some internal server error")
	mockService.On("Delete", mock.Anything, productID).Return(nil, eErr).Once()

	// Act
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{gin.Param{Key: "id", Value: productID}}
	handler.Delete(ctx)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Response should have JSON content type")
	mockService.AssertExpectations(t)

	// Unmarshal and assert response fields
	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "Response body should be valid JSON")
	assert.False(t, resp.Success, "Response should indicate failure")
	assert.Equal(t, "Internal Server Error", resp.Error, "Error message should match with internal server error")
}

func TestProductHandler_ListProducts_Success(t *testing.T) {
	// Arrange
	type testCase struct {
		Name          string
		Query         string
		ExpectedLimit int
		ExpectedPage  int
	}

	testCases := []testCase{
		{Name: "List all products", Query: "", ExpectedLimit: 20, ExpectedPage: 1},
		{Name: "List products with pagination", Query: "?page=1&limit=10", ExpectedLimit: 10, ExpectedPage: 1},
		{Name: "List products with autocorrected query", Query: "?page=0&limit=1000", ExpectedLimit: 100, ExpectedPage: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Arrange
			mockService := &MockProductService{}
			handler := NewProductsHandler(mockService, zap.NewNop())

			req := httptest.NewRequest("GET", "/products"+tc.Query, nil)
			w := httptest.NewRecorder()

			// Mock service expectation
			mockService.On("List", mock.Anything, &models.ListProductsDTO{
				Page:  tc.ExpectedPage,
				Limit: tc.ExpectedLimit,
			}).Return([]models.Product{
				{ID: "1", Name: "Product 1", Description: "Description 1", Price: 100},
				{ID: "2", Name: "Product 2", Description: "Description 2", Price: 200},
			}, 2, nil).Once()

			// Act
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			handler.List(ctx)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 for successful listing")
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Response should have JSON content type")
			mockService.AssertExpectations(t)

			// Unmarshal and assert response fields
			var resp struct {
				Success bool             `json:"success"`
				Data    []models.Product `json:"data"`
				Total   int              `json:"total"`
				Page    int              `json:"page"`
				Size    int              `json:"size"`
				Pages   int              `json:"pages"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err, "Response body should be valid JSON")
			assert.True(t, resp.Success, "Response should indicate success")
			assert.Len(t, resp.Data, 2, "Response data should contain 2 products")
			assert.Equal(t, 2, resp.Total, "Total should be 2")
			assert.Equal(t, tc.ExpectedPage, resp.Page, "Page should match expected")
			assert.Equal(t, tc.ExpectedLimit, resp.Size, "Size should match expected")
			assert.Equal(t, 1, resp.Pages, "Pages should be 1")
		})
	}
}
func TestProductHandler_ListProducts_BadRequest(t *testing.T) {
	// Arrange
	type testCase struct {
		Name  string
		Query string
	}

	testCases := []testCase{
		{Name: "List products with invalid page", Query: "?page=abc"},
		{Name: "List products with invalid limit", Query: "?limit=xyz"},
		{Name: "List products with invalid pagination", Query: "?page=abc&limit=xyz"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Arrange
			mockService := &MockProductService{}
			handler := NewProductsHandler(mockService, zap.NewNop())

			req := httptest.NewRequest("GET", "/products"+tc.Query, nil)
			w := httptest.NewRecorder()

			// Act
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			handler.List(ctx)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for bad request")
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Response should have JSON content type")
			mockService.AssertNotCalled(t, "List")

			// Unmarshal and assert response fields
			var resp struct {
				Success bool   `json:"success"`
				Error   string `json:"error"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err, "Response body should be valid JSON")
			assert.False(t, resp.Success, "Response should indicate failure")
			assert.NotEmpty(t, resp.Error, "Error message should be present")
		})
	}
}

func TestProductHandler_ListProducts_InternalError(t *testing.T) {
	// Arrange
	type testCase struct {
		Name  string
		Query string
	}

	testCases := []testCase{
		{Name: "List products with valid request & internal error", Query: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Arrange
			mockService := &MockProductService{}
			handler := NewProductsHandler(mockService, zap.NewNop())

			req := httptest.NewRequest("GET", "/products"+tc.Query, nil)
			w := httptest.NewRecorder()

			// Simulate internal server error
			mockService.On("List", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("internal error")).Once()

			// Act
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			handler.List(ctx)

			// Assert
			assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected status code 500 for internal server error")
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Response should have JSON content type")
			mockService.AssertExpectations(t)

			// Unmarshal and assert response fields
			var resp struct {
				Success bool   `json:"success"`
				Error   string `json:"error"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err, "Response body should be valid JSON")
			assert.False(t, resp.Success, "Response should indicate failure")
			assert.NotEmpty(t, resp.Error, "Error message should be present")
		})
	}
}
