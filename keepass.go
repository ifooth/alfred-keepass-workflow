package main

import (
	"encoding/json"
	"fmt"
	"os"

	aw "github.com/deanishe/awgo"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/tobischo/gokeepasslib/v3"
)

var client = resty.New()

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

type KeeEntry struct {
	title    string
	group    string
	username string
	password string
}

func (k *KeeEntry) AddItem(wf *aw.Workflow) {
	wf.NewItem(k.title).
		Subtitle(k.username).
		Copytext(k.username).
		Largetype(k.username).
		Var("Password", k.password). // 提供复制内容
		Valid(true)
}

// HTTPGetFile http 方式读取 keepass db
func HTTPGetFile() (*gokeepasslib.DBContent, error) {
	url := GetKeepassURL(cfg)
	resp, err := client.R().SetDoNotParseResponse(true).Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "http get")
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("http resp is not a success, %s", resp.Status())
	}

	password := GetKesspassPwd(cfg)

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(password)
	if err := gokeepasslib.NewDecoder(resp.RawBody()).Decode(db); err != nil {
		return nil, err
	}

	data, err := json.Marshal(db.Content)
	db.Content.Root.Groups[0].Entries[0].GetTitle()
	fmt.Println(len(data), err)

	db.UnlockProtectedEntries()
	return db.Content, nil
}
