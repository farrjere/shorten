package main

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	_ "log"
	"math/rand"
	"net/http"
	"os"

	"github.com/farrjere/shortener/repo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

//go:embed schema.sql
var ddl string
var queries *repo.Queries

type ShortenRequest struct {
	Longurl string `json:"longurl" form:"longurl"`
	Email   string `json:"email" form:"email"`
}

type ShortenResponse struct {
	Longurl   string `json:"longurl"`
	Email     string `json:"email"`
	Shortcode string `json:"shortcode"`
	Status    string `json:"status"`
	Success   bool   `json:"success"`
}

func setupDb() (*sql.DB, error) {
	ctx := context.Background()
	db, err := sql.Open("sqlite3", "file:urls.db")

	if err != nil {
		return db, err
	}

	// create tables
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return db, nil
	}

	return db, nil
}

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func registerUrl(ctx echo.Context) error {
	log.Info("Registering url")
	req := new(ShortenRequest)
	if err := ctx.Bind(req); err != nil {
		return err
	}
	mapping, err := queries.GetUrlMappingByLongurl(ctx.Request().Context(), req.Longurl)
	if err == nil {
		return shortenResp(ctx, mapping, false, "A mapping already existed for that url, see: "+os.Getenv("server_url")+"/"+mapping.Shortcode)
	}
	for {

		short := RandString(4)
		_, err := queries.GetUrlMappingByShortcode(ctx.Request().Context(), short)
		if errors.Is(err, sql.ErrNoRows) {
			mapping := repo.CreateUrlMappingParams{Longurl: req.Longurl, Owner: req.Email, Shortcode: short}
			created, err := queries.CreateUrlMapping(ctx.Request().Context(), mapping)
			success := err == nil
			status := ""
			if err != nil {
				status = err.Error()
			}
			return shortenResp(ctx, created, success, status)
		}
	}
}

func shortenResp(ctx echo.Context, mapping repo.UrlMapping, success bool, status string) error {
	resp := ShortenResponse{Longurl: mapping.Longurl, Email: mapping.Owner, Shortcode: mapping.Shortcode, Success: success, Status: status}
	if len(ctx.FormValue("longurl")) > 0 {
		log.Info("returning an ajax response")
		resp.Shortcode = os.Getenv("server_url") + "/" + resp.Shortcode
		component := RegistrationResponse(resp)
		r := ctx.Response()
		r.Status = http.StatusCreated
		component.Render(ctx.Request().Context(), r.Writer)
		return nil
	} else {
		log.Info("returning some json")
		return ctx.JSON(http.StatusCreated, resp)
	}
}

func renderHome(ctx echo.Context) error {
	component := Page(RegistrationForm())
	r := ctx.Response()
	r.Status = http.StatusOK
	component.Render(ctx.Request().Context(), r.Writer)
	return nil
}

func renderReset(ctx echo.Context) error {
	component := RegistrationForm()
	r := ctx.Response()
	r.Status = http.StatusOK
	component.Render(ctx.Request().Context(), r.Writer)
	return nil
}

func redirectShort(ctx echo.Context) error {
	short := ctx.Param("short")

	mapping, err := queries.GetUrlMappingByShortcode(ctx.Request().Context(), short)
	if err != nil {
		return ctx.String(http.StatusNotFound, "No url registered for that shortcode")
	}
	return ctx.Redirect(http.StatusFound, mapping.Longurl)
}

func main() {
	db, err := setupDb()
	queries = repo.New(db)
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.Static("/static", "assets")
	e.GET("/index.html", renderHome)
	e.GET("/reset", renderReset)
	e.GET("/:short", redirectShort)
	e.POST("/register", registerUrl)
	e.Logger.Fatal(e.Start(":1323"))
}
