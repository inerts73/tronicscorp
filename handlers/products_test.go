package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/inerts73/tronicscorp/config"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	c   *mongo.Client
	db  *mongo.Database
	col *mongo.Collection
	cfg config.Properties
	h   ProductHandler
)

func init() {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Configuration cannot be read : %v", err)
	}
	connectURI := fmt.Sprintf("mongodb://%s:%s", cfg.DBHost, cfg.DBPort)
	c, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connectURI))
	if err != nil {
		log.Fatal("Unable tp connect to database: %w", err)
	}
	db = c.Database(cfg.DBName)
	col = db.Collection(cfg.CollectionName)
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	//set up
	testCode := m.Run()
	//destory
	col.Drop(ctx)
	db.Drop(ctx)
	os.Exit(testCode)
}

func TestProduct(t *testing.T) {
	var docID string

	t.Run("test create product", func(t *testing.T) {
		var IDs []string
		body := `
		[
			{
				"product_name":"googletalk",
				"price":250,
				"currency":"INR",
				"vendor":"google",
				"accessories":["charger","subscription"]
			}
		]
		`
		req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		c := e.NewContext(req, res)
		h.Col = col
		err := h.CreateProducts(c)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, res.Code)

		err = json.Unmarshal(res.Body.Bytes(), &IDs)
		assert.Nil(t, err)
		docID = IDs[0] // assign the value to docID
		t.Logf("IDs: %#+v\n", IDs)
		for _, ID := range IDs {
			assert.NotNil(t, ID)
		}
	})

	t.Run("get products", func(t *testing.T){
		var products []Product
		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		c := e.NewContext(req, res)
		h.Col = col
		err := h.GetProducts(c)
		assert.Nil(t, err)		
		assert.Equal(t, http.StatusOK, res.Code)

		err = json.Unmarshal(res.Body.Bytes(), &products)
		assert.Nil(t, err)
		for _, product := range products {
			assert.Equal(t, "googletalk", product.Name)
		}
	})

	t.Run("get products with query params", func(t *testing.T){
		var products []Product
		req := httptest.NewRequest(http.MethodGet, "/products?currency=INR&vendor=google", nil)
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		c := e.NewContext(req, res)
		h.Col = col
		err := h.GetProducts(c)
		assert.Nil(t, err)		
		assert.Equal(t, http.StatusOK, res.Code)

		err = json.Unmarshal(res.Body.Bytes(), &products)
		assert.Nil(t, err)
		for _, product := range products {
			assert.Equal(t, "googletalk", product.Name)
		}
	})	

	t.Run("get a product", func(t *testing.T){
		var product Product
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/products/%s", docID), nil)
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		c := e.NewContext(req, res)
		c.SetParamNames("id") 					  // this is a hack
		c.SetParamValues(fmt.Sprintf("%s", docID)) //this is a hack
		h.Col = col
		err := h.GetProduct(c)
		assert.Nil(t, err)		
		assert.Equal(t, http.StatusOK, res.Code)
		err = json.Unmarshal(res.Body.Bytes(), &product)
		assert.Nil(t, err)
		assert.Equal(t, "INR", product.Currency)
	})	

	t.Run("put product", func(t *testing.T) {
		var product Product
		body := `
		{
			"product_name":"googletalk",
			"price":250,
			"currency":"USD",
			"vendor":"google",
			"accessories":["charger","subscription"]
		}
		`
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/products/%s", docID), strings.NewReader(body))
		res := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		e := echo.New()
		c := e.NewContext(req, res)
		c.SetParamNames("id")					  // this is a hack	
		c.SetParamValues(fmt.Sprintf("%s", docID)) //this is a hack
		h.Col = col
		err := h.UpdateProduct(c)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code)
		err = json.Unmarshal(res.Body.Bytes(), &product)
		assert.Nil(t, err)
		assert.Equal(t, "USD", product.Currency)
	})	

	t.Run("delete a product", func(t *testing.T) {
		var delCount int64

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/products/%s", docID), nil)
		res := httptest.NewRecorder()
		e := echo.New()
		c := e.NewContext(req, res)
		c.SetParamNames("id")					  // this is a hack	
		c.SetParamValues(fmt.Sprintf("%s", docID)) //this is a hack
		h.Col = col
		err := h.DeleteProduct(c)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code)
		err = json.Unmarshal(res.Body.Bytes(), &delCount)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), delCount)
	})
}
