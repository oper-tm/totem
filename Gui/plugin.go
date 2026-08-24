package main

// The receiver component of the plugin was built by the Claude AI.
// I just didn't have the energy for it anymore :(

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dop251/goja"
)

var (
	pluginTimeout = 60 * time.Second
)

var (
	plName string
)

func RunPlugin(name string, args map[string]interface{}) error {
	src, err := loadPluginSource(name)
	if err != nil {
		return err
	}

	plName = name

	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	host := vm.NewObject()
	registerHostFunctions(vm, host, name)
	vm.Set("host", host)

	if args != nil {
		vm.Set("args", args)
	}

	ctx, cancel := context.WithTimeout(context.Background(), pluginTimeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("plugin %q panicked: %v", name, r)
			}
		}()

		_, runErr := vm.RunString(src)
		done <- runErr
	}()

	select {
	case <-ctx.Done():
		vm.Interrupt("Plugin Timeout")
		app.Event.Emit("lp", "f")
		return fmt.Errorf("plugin %q timed out after %s", name, pluginTimeout)

	case err := <-done:
		if err != nil {
			app.Event.Emit("lp", "f")
			return fmt.Errorf("plugin %q error: %w", name, err)
		}
		return nil
	}
}

func loadPluginSource(name string) (string, error) {
	if strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return "", fmt.Errorf("invalid plugin name: %q", name)
	}

	path := filepath.Join(dp, name+".js")
	b, err := os.ReadFile(path)
	if err != nil {
		app.Event.Emit("lp", "f")
		return "", fmt.Errorf("cannot read plugin %q: %w", name, err)
	}
	return string(b), nil
}

func registerHostFunctions(vm *goja.Runtime, host *goja.Object, pluginName string) {

	host.Set("log", func(msg string) {
		fmt.Printf("[plugin:%s] %s\n", pluginName, msg)
	})

	host.Set("loadingPage", func(s bool) {
		if s {
			app.Event.Emit("lp", "t")
		} else {
			app.Event.Emit("lp", "f")
		}
	})

	host.Set("setInp", func(url string) {
		app.Event.Emit("fuinp", url)
	})

	host.Set("getUrl", func(url, h string) string {
		err1, err2, data := getFile(url, plName, h, false)

		logger.Println(err1, err2)
		return string(data)
	})

	host.Set("getUrlPost", func(a, b string, c, d map[string]interface{}, e bool) string {
		_, _, data := getFilePost(a, b, c, d, e)
		return string(data)
	})

	host.Set("base64", func(url string) string {
		urlD := base64.RawURLEncoding.EncodeToString([]byte(url))
		return string(urlD)
	})

	host.Set("II_A", func(title, desc, onclick string, ri bool) {
		if ri {
			app.Event.Emit("ihrem")
		}
		app.Event.Emit("ih", `<div class="itemA" onclick="`+onclick+`">
			<a class="title">`+title+`</a>
			<a class="desc">`+desc+`</a>
		</div>`)
	})

	host.Set("II_B", func(poster, title, desc, onclick string, ri bool) {
		if ri {
			app.Event.Emit("ihrem")
		}
		app.Event.Emit("ih", `<div class="itemB" onclick="`+onclick+`">
			<img src="/photo/`+poster+`" alt="" class="poster">
			<a class="title">`+title+`</a>
			<a class="desc">`+desc+`</a>
		</div>`)
	})
}

func getPluginList() []string {
	entries, _ := os.ReadDir(dp)

	var files []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) == ".js" {
			files = append(files, entry.Name())
		}
	}

	fmt.Println(files)

	return files
}
