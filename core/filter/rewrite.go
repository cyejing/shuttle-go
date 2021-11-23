package filter

import (
	"regexp"
)

type rewrite struct {
	name string
}

func init() {
	RegistryFilter(&rewrite{name: "rewrite"})
}

type RewriteConfig struct {
	Regex       string
	Replacement string
}

func (r rewrite) Init() {
}

func (r rewrite) Name() string {
	return r.name
}

func (r rewrite) Filter(exchange *Exchange, c interface{}) error {
	var config RewriteConfig
	if err := mapstruct(c, &config); err != nil {
		return err
	}

	compile, err := regexp.Compile(config.Regex)
	if err != nil {
		return err
	}
	oldPath := exchange.Req.URL.Path
	newPath := compile.ReplaceAllString(oldPath, config.Replacement)
	exchange.Req.RequestURI = newPath
	exchange.Req.URL.Path = newPath

	log.Debugf("rewrite path. oldPath:%v newPtah:%v", oldPath, newPath)

	return nil
}
