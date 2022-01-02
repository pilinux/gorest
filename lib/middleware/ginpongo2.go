package middleware

// github.com/ngerakines/ginpongo2
// The MIT License (MIT)
// Copyright (c) 2014 Nick Gerakines
// Copyright (c) 2022 piLinux

import (
	"net/http"

	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Pongo2 - uses the Pongo2 template library
// https://github.com/flosch/pongo2
// to render templates
func Pongo2() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		name := stringFromContext(c, "template")
		data, _ := c.Get("data")

		if name == "" {
			return
		}

		// Set base directory
		fs, err := pongo2.NewLocalFileSystemLoader("")
		if err != nil {
			log.WithError(err).Panic("panic code: 300")
		}

		s := pongo2.NewSet("set base directory", fs)
		s.Globals["base_directory"] = "templates/"

		if err := fs.SetBaseDir(s.Globals["base_directory"].(string)); err != nil {
			log.WithError(err).Panic("panic code: 301")
		}

		template, err := s.FromFile(name)
		if err != nil {
			log.WithError(err).Panic("panic code: 302")
		}

		err = template.ExecuteWriter(convertContext(data), c.Writer)
		if err != nil {
			log.WithError(err).Panic("panic code: 303")
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		}
	}
}

// stringFromContext ...
func stringFromContext(c *gin.Context, input string) string {
	raw, ok := c.Get(input)
	if ok {
		strVal, ok := raw.(string)
		if ok {
			return strVal
		}
	}
	return ""
}

// convertContext ...
func convertContext(thing interface{}) pongo2.Context {
	if thing != nil {
		context, isMap := thing.(map[string]interface{})
		if isMap {
			return context
		}
	}
	return nil
}
