package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails"
)

type Todos struct {
	filename string
	runtime  *wails.Runtime
	logger   *wails.CustomLogger
	watcher  *fsnotify.Watcher
}

func (t *Todos) WailsInit(runtime *wails.Runtime) error {
	t.runtime = runtime
	t.logger = t.runtime.Log.New("Todos")
	t.logger.Info("I'm here")

	// Set the default filename to ~/mylist.json
	homedir, err := runtime.FileSystem.HomeDir()
	if err != nil {
		return err
	}
	t.filename = path.Join(homedir, "mylist.json")
	t.runtime.Window.SetTitle(t.filename)
	t.ensureFileExists()
	return t.startWatcher()
}

func (t *Todos) ensureFileExists() {
	// Check status of file
	_, err := os.Stat(t.filename)
	// If it doesn't exist
	if os.IsNotExist(err) {
		// Create a default file
		ioutil.WriteFile(t.filename, []byte("[]"), 0600)
	}
}

func (t *Todos) setFilename(filename string) error {
	var err error
	// Stop watching the current file and return any error
	err = t.watcher.Remove(t.filename)
	if err != nil {
		return err
	}

	// Set filename
	t.filename = filename

	// Add the new file to the watcher and return any errors
	err = t.watcher.Add(filename)
	if err != nil {
		return err
	}
	t.logger.Info("Now watching: " + filename)
	t.runtime.Window.SetTitle(filename)
	return nil
}

func (t *Todos) saveListByName(todos string, filename string) error {
	return ioutil.WriteFile(filename, []byte(todos), 0600)
}

func (t *Todos) LoadNewList() {
	filename := t.runtime.Dialog.SelectFile()
	if len(filename) > 0 {
		t.setFilename(filename)
		t.runtime.Events.Emit("filemodified")
	}
}

func (t *Todos) startWatcher() error {
	t.logger.Info("Starting watcher!")
	watcher, err := fsnotify.NewWatcher()
	t.watcher = watcher
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					t.logger.Infof("modified file: %s", event.Name)
					t.runtime.Events.Emit("filemodified")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				t.logger.Error(err.Error())
			}
		}
	}()

	err = watcher.Add(t.filename)
	if err != nil {
		return err
	}
	return nil
}

// NewTodos attempts to create a new Todo list
func NewTodos() (*Todos, error) {
	// Create todos instance
	result := &Todos{}
	// Get cwd
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// Join the cwd with filename
	filename := path.Join(cwd, "mylist.json")
	// Set the filename member of our new todo list
	result.filename = filename
	return result, nil
}

func (t *Todos) LoadList() (string, error) {
	t.logger.Infof("Loading list from: %s", t.filename)
	bytes, err := ioutil.ReadFile(t.filename)
	if err != nil {
		err = fmt.Errorf("Unable to open list: %s", t.filename)
	}
	t.runtime.Events.Emit("error", "I am from go!", 1234)
	return string(bytes), err
}

func (t *Todos) SaveList(todos string) error {
	t.logger.Infof("Saving list: %s", todos)
	return t.saveListByName(todos, t.filename)
}

func (t *Todos) SaveAs(todos string) error {
	filename := t.runtime.Dialog.SelectSaveFile()
	t.logger.Info("Save As: " + filename)
	err := t.saveListByName(todos, filename)
	if err != nil {
		return err
	}
	return t.setFilename(filename)
}
