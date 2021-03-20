package api_test

import (
	"api"
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	bytes := make([]rune, n)

	for i := range bytes {
		bytes[i] = letters[rand.Intn(len(letters))]
	}

	return string(bytes)
}

func eq(one, another api.CarModel) bool {
	return one.Id == another.Id &&
		one.Brand == another.Brand &&
		one.Model == another.Model &&
		one.Price == another.Price &&
		one.Status == another.Status &&
		one.Mileage == another.Mileage
}

type mockDB struct {
	CarModels []*api.CarModel
	LastID    api.IDType
}

func (db *mockDB) find(id api.IDType) (int, *api.CarModel, error) {
	for idx, cm := range db.CarModels {
		if cm.Id == id {
			return idx, cm, nil
		}
	}

	return -1, nil, api.ErrNotFound
}

func (db *mockDB) push(brand, model string, price int64, status api.Status, mileage int64) *api.CarModel {
	cm := new(api.CarModel)

	cm.Id = db.LastID
	cm.Brand = brand
	cm.Model = model
	cm.Price = price
	cm.Status = status
	cm.Mileage = mileage

	db.CarModels = append(db.CarModels, cm)
	db.LastID++

	return cm
}

func (db *mockDB) List(ctx context.Context) ([]*api.CarModel, error) {
	cms := db.CarModels
	return cms, nil
}

func (db *mockDB) Get(ctx context.Context, id api.IDType) (*api.CarModel, error) {
	_, cm, err := db.find(id)
	return cm, err
}

func (db *mockDB) Create(ctx context.Context, brand, model string, price int64, status api.Status, mileage int64) (*api.CarModel, error) {
	if price < 0 {
		return nil, api.ErrBadPrice
	}
	if mileage < 0 {
		return nil, api.ErrBadMileage
	}

	cm := db.push(brand, model, price, status, mileage)
	return cm, nil
}

func (db *mockDB) Update(ctx context.Context, id api.IDType, brand, model string, price int64, status api.Status, mileage int64) (*api.CarModel, error) {
	if price < 0 {
		return nil, api.ErrBadPrice
	}
	if mileage < 0 {
		return nil, api.ErrBadMileage
	}

	_, cm, err := db.find(id)
	if err != nil {
		return nil, err
	}

	cm.Brand = brand
	cm.Model = model
	cm.Price = price
	cm.Status = status
	cm.Mileage = mileage

	return cm, nil
}

func (db *mockDB) Delete(ctx context.Context, id api.IDType) error {
	idx, _, err := db.find(id)
	if err != nil {
		return err
	}

	cms := db.CarModels
	cms[len(cms)-1], cms[idx] = cms[idx], cms[len(cms)-1]
	db.CarModels = cms[:len(cms)-1]

	return nil
}

func (db *mockDB) Close(ctx context.Context) error { return nil }

func setupDB() *mockDB {
	db := new(mockDB)
	for i := 0; i < 25; i++ {
		db.push(
			randString(16),
			randString(16),
			int64(rand.Intn(256)*i*1000),
			api.Status(rand.Intn(4)),
			int64(rand.Intn(256)),
		)
	}

	return db
}

func setupServer(db *mockDB) *api.Server {
	server := &api.Server{DB: db}
	return server
}

func TestServer_List(t *testing.T) {
	db := setupDB()
	server := setupServer(db)

	request, _ := http.NewRequest(http.MethodGet, "/tt/v0/cars", nil)
	response := httptest.NewRecorder()

	server.List()(response, request)

	if got, want := response.Code, http.StatusOK; got != want {
		t.Fatalf("Unexpected code: got %d, want %d\nResponse is '%s'", got, want, response.Body)
	}

	bytes := response.Body
	cms := new([]*api.CarModel)
	err := json.Unmarshal(bytes.Bytes(), cms)
	if err != nil {
		t.Fatal(err)
	}
}

func TestServer_Get(t *testing.T) {
	db := setupDB()
	server := setupServer(db)
	id := 0

	request, _ := http.NewRequest(
		http.MethodGet,
		strings.Join([]string{"/tt/v0/cars", strconv.FormatInt(int64(id), 10)}, "/"),
		nil,
	)
	response := httptest.NewRecorder()
	// gorilla/mux context problem
	// see: https://stackoverflow.com/questions/34435185/unit-testing-for-functions-that-use-gorilla-mux-url-parameters
	request = mux.SetURLVars(request, map[string]string{
		"id": strconv.FormatInt(int64(id), 10),
	})

	server.Get()(response, request)

	if got, want := response.Code, http.StatusOK; got != want {
		t.Fatalf("Unexpected code: got %d, want %d\nResponse is %s", got, want, response.Body)
	}

	bytes := response.Body
	cm := new(api.CarModel)
	err := json.Unmarshal(bytes.Bytes(), cm)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := cm, db.CarModels[id]; !eq(*got, *want) {
		t.Fatalf("Unexpected item: got %v, want %v", got, want)
	}
}

func TestServer_Create(t *testing.T) {
	db := setupDB()
	server := setupServer(db)
	newcm := `{"brand": "brand", "model": "model", "price": 123456, "status": "В пути", "mileage": 123456}`

	request, _ := http.NewRequest(
		http.MethodPost,
		"/tt/v0/cars",
		strings.NewReader(newcm),
	)
	response := httptest.NewRecorder()
	server.Create()(response, request)

	if got, want := response.Code, http.StatusAccepted; got != want {
		t.Fatalf("Unexpected code: got %d, want %d\nResponse is %s", got, want, response.Body)
	}

	bytes := response.Body
	cm := new(api.CarModel)
	err := json.Unmarshal(bytes.Bytes(), cm)
	if err != nil {
		t.Fatal(err)
	}
}

func TestServer_Update(t *testing.T) {
	db := setupDB()
	server := setupServer(db)
	id := 0
	newcm := `{"brand": "brand", "model": "model", "price": 123456, "status": "В пути", "mileage": 123456}`

	request, _ := http.NewRequest(
		http.MethodPut,
		strings.Join([]string{"/tt/v0/cars", strconv.FormatInt(int64(id), 10)}, "/"),
		strings.NewReader(newcm),
	)
	response := httptest.NewRecorder()
	// gorilla/mux Vars from context problem
	// see: https://stackoverflow.com/questions/34435185/unit-testing-for-functions-that-use-gorilla-mux-url-parameters
	request = mux.SetURLVars(request, map[string]string{
		"id": strconv.FormatInt(int64(id), 10),
	})
	server.Update()(response, request)
	if got, want := response.Code, http.StatusAccepted; got != want {
		t.Fatalf("Unexpected code: got %d, want %d\nResponse is %s", got, want, response.Body)
	}

	bytes := response.Body
	cm := new(api.CarModel)
	err := json.Unmarshal(bytes.Bytes(), cm)
	if err != nil {
		t.Fatal(err)
	}
}

func TestServer_Delete(t *testing.T) {
	db := setupDB()
	server := setupServer(db)
	id := 0

	request, _ := http.NewRequest(
		http.MethodDelete,
		strings.Join([]string{"/tt/v0/cars", strconv.FormatInt(int64(id), 10)}, "/"),
		nil,
	)
	response := httptest.NewRecorder()
	// gorilla/mux context problem
	// see: https://stackoverflow.com/questions/34435185/unit-testing-for-functions-that-use-gorilla-mux-url-parameters
	request = mux.SetURLVars(request, map[string]string{
		"id": strconv.FormatInt(int64(id), 10),
	})
	server.Delete()(response, request)
	if got, want := response.Code, http.StatusNoContent; got != want {
		t.Fatalf("Unexpected code: got %d, want %d\nResponse is %s", got, want, response.Body)
	}
}
