package api_test

import (
	"api"
	"context"
	"fmt"
	"github.com/caarlos0/env"
	"math/rand"
	"testing"
)

type dbconfig struct {
	DbHost     string `env:"DB_HOST" envDefault:"localhost"`
	DbPort     int    `env:"DB_PORT" envDefault:"5432"`
	DbUser     string `env:"POSTGRES_USER" envDefault:"user"`
	DbPassword string `env:"POSTGRES_PASSWORD" envDefault:"password"`
	DbDatabase string `env:"POSTGRES_DB"`
}

func formConnString() (string, error) {
	cfg := dbconfig{}
	if err := env.Parse(&cfg); err != nil {
		return "", err
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.DbUser, cfg.DbPassword, cfg.DbHost, cfg.DbPort, cfg.DbDatabase), nil
}

type HalfBakedCarModel struct {
	Brand   string     `json:"brand"`
	Model   string     `json:"model"`
	Price   int64      `json:"price"`
	Status  api.Status `json:"status"`
	Mileage int64      `json:"mileage"`
}

func generateHBCM(idx int) HalfBakedCarModel {
	return HalfBakedCarModel{
		Brand:   randString(256),
		Model:   randString(256),
		Price:   int64(rand.Intn(256 * idx)),
		Status:  api.Status(0),
		Mileage: int64(rand.Intn(256 * idx * 1000)),
	}
}

func TestPSQLdb_All(t *testing.T) {
	connString, err := formConnString()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	db, err := api.NewDB(ctx, connString)
	if err != nil {
		t.Fatal(err)
	}

	count := 10
	cm_s := make([]*api.CarModel, 0)
	// Create
	for i := 0; i < count; i++ {
		hbcm := generateHBCM(i + 1)
		cm, err := db.Create(ctx, hbcm.Brand, hbcm.Model, hbcm.Price, hbcm.Status, hbcm.Mileage)
		if err != nil {
			t.Fatal(err)
		}

		cm_s = append(cm_s, cm)
	}
	// List
	_, err = db.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Get
	for i := 0; i < count; i++ {
		cm := cm_s[i]
		_, err = db.Get(ctx, cm.Id)
		if err != nil {
			t.Fatal(err)
		}
	}
	// Update
	for i := 0; i < count; i++ {
		cm := cm_s[i]
		_, err = db.Update(ctx, cm.Id, "pupa", "lupa", 100, api.Status(0), 100)
		if err != nil {
			t.Fatal()
		}
	}
	// Delete
	for i := 0; i < count; i++ {
		cm := cm_s[i]
		err = db.Delete(ctx, cm.Id)
		if err != nil {
			t.Fatal(err)
		}
	}
}
