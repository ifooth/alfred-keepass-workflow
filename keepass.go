package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	cacheKey    = "cached_kdbx.dat"
	titleKey    = "Title"
	usernameKey = "UserName"
	passwordKey = "Password"
	urlKey      = "URL"
	notesKey    = "Notes"
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
	dbPath        string
	password      string
	Entries       []*KeeEntry `json:"-"`
	EncryptedData []byte      `json:"encrypted_data"`
	LastModified  time.Time   `json:"last_modified"`
}

// NewKee
func NewKee(dbPath, password string) *Kee {
	return &Kee{
		dbPath:   dbPath,
		password: password,
		Entries:  []*KeeEntry{},
	}
}

// CheckUpdate
func (k *Kee) CheckDBUpdate() (time.Time, error) {
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

// LoadEntries
func (k *Kee) LoadEntries() error {
	st := time.Now()
	defer func() {
		log.Printf("load entries durations: %s", time.Since(st))
	}()

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(k.password)
	if err := gokeepasslib.NewDecoder(bytes.NewBuffer(k.EncryptedData)).Decode(db); err != nil {
		return err
	}
	db.UnlockProtectedEntries()

	if len(db.Content.Root.Groups) > 0 {
		k.LoadGroups(db.Content.Root.Groups)
	}

	return nil
}

// LoadGroups 递归查询
func (k *Kee) LoadGroups(groups []gokeepasslib.Group) {
	for _, group := range groups {
		for _, entry := range group.Entries {
			k.Entries = append(k.Entries, &KeeEntry{
				Title:    entry.GetTitle(),
				Group:    group.Name,
				Username: entry.GetContent(usernameKey),
				Password: entry.GetPassword(),
			})
		}
		if len(group.Groups) > 0 {
			k.LoadGroups(group.Groups)
		}
	}
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

	data, err := ioutil.ReadAll(resp.RawBody())
	if err != nil {
		return err
	}

	k.LastModified = t
	k.EncryptedData = data
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
	} else {
		log.Printf("load from cache: %s", cacheKey)
	}

	lastModified, err := k.CheckDBUpdate()
	if err != nil {
		log.Printf("check db update err: %s", err)
	}
	if lastModified.After(k.LastModified) {
		if err := k.LoadAndCache(); err != nil {
			wf.FatalError(err)
			return
		}
	}

	if err := k.LoadEntries(); err != nil {
		wf.FatalError(err)
		return
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
		Var("username", k.Username). // 提供复制内容
		Var("password", k.Password).
		Valid(true)
}
