package api

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrNotFound   = errors.New("car model not found")
	ErrBadPrice   = errors.New("'price' field is invalid")
	ErrBadMileage = errors.New("'mileage' field is invalid")
)

type IDType int

type Status int8

const (
	ontheway Status = iota
	stored
	sold
	withdrawnfromsale
)

var statusToString = map[Status]string{
	ontheway:          "В пути",
	stored:            "На складе",
	sold:              "Продан",
	withdrawnfromsale: "Снят с продажи",
}
var stringToStatus = map[string]Status{
	"В пути":         ontheway,
	"На складе":      stored,
	"Продан":         sold,
	"Снят с продажи": withdrawnfromsale,
}

func (s *Status) Scan(value interface{}) error {
	asString, ok := value.(string)
	if !ok {
		return errors.New("Scan source is not string")
	}
	*s = stringToStatus[asString]
	return nil
}

func (s Status) Value() (driver.Value, error) {
	if value, ok := statusToString[s]; !ok {
		return nil, errors.New("Wrong value for Status")
	} else {
		return value, nil
	}
}

func (s Status) Stringify() string {
	return statusToString[s]
}

func (s Status) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(statusToString[s])
	buffer.WriteString(`"`)

	return buffer.Bytes(), nil
}

func (s *Status) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	*s = stringToStatus[str]
	return nil
}

//1. Уникальный идентификатор (любой тип, общение с БД не является критерием чего-либо, можно сделать и in-memory хранилище на время жизни сервиса)
//2. Бренд автомобиля (текст)
//3. Модель автомобиля (текст)
//4. Цена автомобиля (целое, не может быть меньше 0)
//5. Статус автомобиля (В пути, На складе, Продан, Снят с продажи)
//6. Пробег (целое)
type CarModel struct {
	Id      IDType `json:"id"`
	Brand   string `json:"brand"`
	Model   string `json:"model"`
	Price   int64  `json:"price"`
	Status  Status `json:"status"`
	Mileage int64  `json:"mileage"`
}

type DB interface {
	List(ctx context.Context) ([]*CarModel, error)
	Get(ctx context.Context, id IDType) (*CarModel, error)
	Create(ctx context.Context, brand, model string, price int64, status Status, mileage int64) (*CarModel, error)
	Update(ctx context.Context, id IDType, brand, model string, price int64, status Status, mileage int64) (*CarModel, error)
	Delete(ctx context.Context, id IDType) error
	Close(ctx context.Context) error
}

type PSQLdb struct {
	*pgxpool.Pool
}

func NewDB(ctx context.Context, connString string) (*PSQLdb, error) {
	pool, err := pgxpool.Connect(ctx, connString)
	return &PSQLdb{pool}, err
}

func (db *PSQLdb) List(ctx context.Context) ([]*CarModel, error) {
	query := "SELECT id, brand, model, price, status, mileage FROM CarModel"
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	cms := make([]*CarModel, 0)
	for rows.Next() {
		cm := new(CarModel)
		err := rows.Scan(&cm.Id, &cm.Brand, &cm.Model, &cm.Price, &cm.Status, &cm.Mileage)
		if err != nil {
			return nil, err
		}

		cms = append(cms, cm)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cms, nil

}

func (db *PSQLdb) Get(ctx context.Context, id IDType) (*CarModel, error) {
	query := "SELECT id, brand, model, price, status, mileage FROM CarModel WHERE id=$1"
	row := db.QueryRow(ctx, query, id)
	cm := new(CarModel)
	err := row.Scan(&cm.Id, &cm.Brand, &cm.Model, &cm.Price, &cm.Status, &cm.Mileage)
	if err != nil {
		return nil, err
	}

	return cm, nil
}

func (db *PSQLdb) Create(ctx context.Context, brand, model string, price int64, status Status, mileage int64) (*CarModel, error) {
	if price < 0 {
		return nil, ErrBadPrice
	}
	if mileage < 0 {
		return nil, ErrBadMileage
	}

	query := "INSERT INTO CarModel (brand, model, price, status, mileage) " +
		"VALUES ($1, $2, $3, $4, $5) " +
		"RETURNING id"
	var id IDType
	err := db.QueryRow(ctx, query, brand, model, price, status, mileage).Scan(&id)
	if err != nil {
		return nil, err
	}

	cm := new(CarModel)
	cm.Id = id
	cm.Model = model
	cm.Brand = brand
	cm.Price = price
	cm.Status = status
	cm.Mileage = mileage

	return cm, nil
}

func (db *PSQLdb) Update(ctx context.Context, id IDType, brand, model string, price int64, status Status, mileage int64) (*CarModel, error) {
	if price < 0 {
		return nil, ErrBadPrice
	}
	if mileage < 0 {
		return nil, ErrBadMileage
	}

	query := "UPDATE CarModel SET brand=$1, model=$2, price=$3, status=$4, mileage=$5 WHERE id=$6"

	result, err := db.Exec(ctx, query, brand, model, price, status, mileage, id)
	if err != nil {
		return nil, err
	}
	nRows := result.RowsAffected()
	if nRows == 0 {
		return nil, ErrNotFound
	}

	cm := new(CarModel)
	cm.Id = id
	cm.Brand = brand
	cm.Model = model
	cm.Price = price
	cm.Status = status
	cm.Mileage = mileage

	return cm, nil
}

func (db *PSQLdb) Delete(ctx context.Context, id IDType) error {
	query := "DELETE FROM CarModel WHERE id=$1"
	result, err := db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	nRows := result.RowsAffected()
	if nRows == 0 {
		return ErrNotFound
	}

	return nil
}

func (db *PSQLdb) Close(ctx context.Context) error {
	db.Pool.Close()
	return nil
}
