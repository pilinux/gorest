package middleware

// github.com/ngerakines/ginpongo2
// The MIT License (MIT)
// Copyright (c) 2014 Nick Gerakines
// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import (
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Pongo2 uses the Pongo2 template library
// https://github.com/flosch/pongo2
// to render templates
//
// Example: baseDirectory = "templates/"
func Pongo2(baseDirectory string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Automatic recovery from panic
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic msg: middleware -> pongo2 panicked")
			}
		}()

		c.Next()

		name := StringFromContext(c, "template")
		data, exists := c.Get("data")

		if name == "" {
			return
		}
		if !exists {
			return
		}

		// Set base directory
		fs, err := pongo2.NewLocalFileSystemLoader("")
		if err != nil {
			log.WithError(err).Error("error msg: middleware -> pongo2 failed to create a new LocalFilesystemLoader")
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}

		s := pongo2.NewSet("set base directory", fs)
		s.Globals["base_directory"] = baseDirectory

		if err := fs.SetBaseDir(s.Globals["base_directory"].(string)); err != nil {
			log.WithError(err).Error("error msg: middleware -> pongo2 failed to set base directory")
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}

		template, err := s.FromFile(name)
		if err != nil {
			log.WithError(err).Error("error msg: middleware -> pongo2 base directory not found")
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}

		err = template.ExecuteWriter(ConvertContext(data), c.Writer)
		if err != nil {
			log.WithError(err).Error("error msg: middleware -> pongo2 failed to execute the template with the given context")
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// StringFromContext function retrieves the value from the context and returns it as a string
func StringFromContext(c *gin.Context, input string) string {
	raw, ok := c.Get(input)
	if ok {
		strVal, ok := raw.(string)
		if ok {
			return strVal
		}
	}
	return ""
}

// ConvertContext function converts the input map to a pongo2.Context type and preserves the key-value pairs
func ConvertContext(thing interface{}) pongo2.Context {
	if thing != nil {
		context, isMap := thing.(map[string]interface{})
		if isMap {
			return context
		}
	}
	return nil
}
