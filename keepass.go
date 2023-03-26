package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/tobischo/gokeepasslib/v3"
)

var client = resty.New()

const (
	cacheKey = "cached_kdbx.dat"
)

// GetKeepassURL :
func GetKeepassURL(cfg *aw.Config) string {
	url := os.Getenv("KEEPASS_URL")
	if url != "" {
		return url
	}
	if cfg != nil {
		return cfg.GetString("KEEPASS_URL", "")
	}
	return ""

}

// GetKesspassPwd :
func GetKesspassPwd(cfg *aw.Config) string {
	pwd := os.Getenv("KEEPASS_PWD")
	if pwd != "" {
		return pwd
	}

	if cfg != nil {
		return cfg.GetString("KEEPASS_PWD", "")
	}

	return ""
}

// Kee
type Kee struct {
	dbPath       string
	password     string
	Entries      []*KeeEntry `json:"entries"`
	LastModified time.Time   `json:"last_modified"`
}

// NewKee
func NewKee(dbPath, password string) *Kee {
	return &Kee{
		dbPath:   dbPath,
		password: password,
	}
}

// CheckUpdate
func (k *Kee) CheckDBUpdate() (time.Time, error) {
	fmt.Println(k.dbPath, k.password)
	resp, err := client.R().SetDoNotParseResponse(true).Head(k.dbPath)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "http fetch")
	}
	if !resp.IsSuccess() {
		return time.Time{}, fmt.Errorf("http resp is not a success, %s", resp.Status())
	}

	lastModified := resp.Header().Get("Last-Modified")
	if lastModified == "" {
		return time.Time{}, fmt.Errorf("lastModified is empty, %s", resp.Header())
	}

	t, err := http.ParseTime(lastModified)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "http ParseTime")
	}

	return t, nil
}

// LoadOrStore
func (k *Kee) LoadAndCache() error {
	resp, err := client.R().SetDoNotParseResponse(true).Get(k.dbPath)
	if err != nil {
		return errors.Wrap(err, "http fetch")
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("http resp is not a success, %s", resp.Status())
	}

	lastModified := resp.Header().Get("Last-Modified")
	if lastModified == "" {
		return fmt.Errorf("lastModified is empty, %s", resp.Header())
	}

	t, err := http.ParseTime(lastModified)
	if err != nil {
		return errors.Wrap(err, "http ParseTime")
	}

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(k.password)
	if err := gokeepasslib.NewDecoder(resp.RawBody()).Decode(db); err != nil {
		return err
	}

	entrys := make([]*KeeEntry, 0)
	for _, group := range db.Content.Root.Groups {
		for _, entry := range group.Entries {
			entrys = append(entrys, &KeeEntry{
				Title:    entry.GetTitle(),
				Group:    group.Name,
				Username: entry.GetContent("Username"),
				Password: entry.GetPassword(),
			})
		}
	}

	k.LastModified = t
	k.Entries = entrys
	wf.Cache.StoreJSON(cacheKey, k)
	return nil
}

// Load
func (k *Kee) Load(wf *aw.Workflow) {
	err := wf.Cache.LoadJSON(cacheKey, k)
	if err != nil {
		if err := k.LoadAndCache(); err != nil {
			wf.FatalError(err)
			return
		}
	}
	lastModified, err := k.CheckDBUpdate()
	if err != nil {
		log.Printf("lastModified err: %s", err)
	}
	if lastModified.Before(k.LastModified) {
		if err := k.LoadAndCache(); err != nil {
			wf.FatalError(err)
			return
		}
	}

	for _, v := range k.Entries {
		v.AddItem(wf)
	}
}

// KeeEntry
type KeeEntry struct {
	Title    string `json:"title"`
	Group    string `json:"group"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// AddItem
func (k *KeeEntry) AddItem(wf *aw.Workflow) {
	wf.NewItem(k.Title).
		Subtitle(k.Username).
		Copytext(k.Username).
		Largetype(k.Username).
		Var("Password", k.Password). // 提供复制内容
		Valid(true)
}
