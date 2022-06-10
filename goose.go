package goose

import (
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	eyeson "github.com/eyeson-team/eyeson-go"
	"github.com/gofiber/template/html"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// videoService provides access to the eyeson API service.
var videoService *eyeson.Client

// db keeps a connection to the platform sqlite database.
var db *gorm.DB

// DATETIME_FORMAT ensures a fixed date time format.
const DATETIME_FORMAT string = "2006-01-02T15:04:05.000Z"

// VARDIR is the storage directory for recordings and snapshots.
const VARDIR string = "/tmp/meetduck"

// AUTHORIZED_DOMAIN defines the email domain that allows users to login.
const AUTHORIZED_DOMAIN string = "eyeson.com"

// NewViewEngine provides a fiber html engine configured to load views from the
// provided views path. It also registers the following template functions:
//
//  - markdown ... {{.Content | markdown}} render markdown content to html
//  - upper ... {{.Name | upper}} upcase a given string
//  - datetime ... {{.CreatedAt | datetime}} output a time element
//
func NewViewEngine(viewsPath string) *html.Engine {
	engine := html.New(viewsPath, ".tmpl")
	engine.AddFunc(
		"markdown", func(args ...interface{}) template.HTML {
			s := blackfriday.Run([]byte(fmt.Sprintf("%s", args...)),
				blackfriday.WithExtensions(blackfriday.HardLineBreak))
			html := bluemonday.UGCPolicy().SanitizeBytes(s)
			return template.HTML(html)
		},
	)
	engine.AddFunc("upper", strings.ToUpper)
	engine.AddFunc("datetime", func(d time.Time) template.HTML {
		s := fmt.Sprintf("<time datetime=\"%s\">%s</time>", d.Format(DATETIME_FORMAT),
			d.Format(time.RFC822))
		return template.HTML(s)
	})
	return engine
}

// InitEyeson creates a eyeson service video client from the given api key.
func InitEyeson(apiKey string) error {
	if len(apiKey) == 0 {
		return errors.New("API_KEY not set")
	}
	videoService = eyeson.NewClient(apiKey)
	return nil
}

// InitDatabase initializes the application database and migrates record
// models.
func InitDatabase(filename string) (err error) {
	db, err = gorm.Open(sqlite.Open(filename), &gorm.Config{})

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Workspace{})
	db.AutoMigrate(&Login{})
	db.AutoMigrate(&Meeting{})
	db.AutoMigrate(&Recording{})
	db.AutoMigrate(&Snapshot{})
	return
}
