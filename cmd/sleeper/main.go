package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// model

type Sleepers struct {
	SleeperUUID string  `gorm:"column:sleeper_uuid" json:"sleeper_uuid"`
	SleeperTime float64 `gorm:"column:sleeper_time" json:"sleeper_time"`
}

type Sleeper_Request struct {
	Number int64 `json:"number"`
}

const DB_PATH = "/Users/neverholiday/sqlitedbs/sleeper.db"

// repositories

type SleeperRepository struct {
	DB *gorm.DB
}

func (r *SleeperRepository) Create(ctx context.Context, sleepers []Sleepers) error {

	err := r.DB.Table("sleepers").WithContext(ctx).Create(sleepers).Error
	if err != nil {
		return fmt.Errorf("create failed, database error: %v", err)
	}
	return nil
}

func (r *SleeperRepository) Delete(ctx context.Context) error {

	err := r.DB.WithContext(ctx).Exec("DELETE FROM sleepers").Error
	if err != nil {
		return fmt.Errorf("delete failed, database error: %v", err)
	}
	return nil
}

func (r *SleeperRepository) Get(ctx context.Context, sleeperUUID string) (Sleepers, error) {

	var sleeper Sleepers
	err := r.DB.Table("sleepers").WithContext(ctx).First(&sleeper, "sleeper_uuid = ?", sleeperUUID).Error
	if err != nil {
		return Sleepers{}, fmt.Errorf("create failed, database error: %v", err)
	}
	return sleeper, nil
}

// app

type App struct {
	SleeperRepo *SleeperRepository
}

func (a *App) Create(c echo.Context) error {

	ctx := c.Request().Context()
	var req Sleeper_Request

	err := c.Bind(&req)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	var sleepers []Sleepers
	for i := 0; i < int(req.Number); i++ {
		sleepers = append(sleepers, Sleepers{
			SleeperUUID: uuid.New().String(),
			SleeperTime: rand.Float64() * 10,
		})
	}

	err = a.SleeperRepo.Create(ctx, sleepers)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, sleepers)
}

func (a *App) Delete(c echo.Context) error {
	ctx := c.Request().Context()

	err := a.SleeperRepo.Delete(ctx)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (a *App) Get(c echo.Context) error {
	ctx := c.Request().Context()

	sleeperUUID := c.Param("sleeper")
	sleeper, err := a.SleeperRepo.Get(ctx, sleeperUUID)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, &sleeper)
}

func main() {
	db, err := gorm.Open(sqlite.Open(DB_PATH), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	sleeperRepo := &SleeperRepository{DB: db}

	app := &App{SleeperRepo: sleeperRepo}

	e := echo.New()
	e.POST("/create", app.Create)
	e.POST("/delete", app.Delete)
	e.GET("/get/:sleeper", app.Get)
	e.Logger.Fatal(e.Start(":8080"))

}
