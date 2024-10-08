package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type PluginsData struct {
	Plugins map[string]*ProcData `json:"plugins"`
}

func NewEmptyPluginsData() *PluginsData {
	plugins := make(map[string]*ProcData)
	plugins["example"] = NewEmptyProcData()
	data := &PluginsData{
		Plugins: plugins,
	}
	return data
}

type ProcData struct {
	Proc string   `json:"proc"`
	Args []string `json:"args"`
}

func NewEmptyProcData() *ProcData {
	proc := &ProcData{
		Proc: "",
		Args: []string{},
	}
	return proc
}

func LoadPlugin() (*PluginsData, error) {
	if !IsExist(PLUGIN_DIR) {
		os.Mkdir(PLUGIN_DIR, 0777)
	}

	path := filepath.Join(PLUGIN_DIR, PLUGIN_JSON)
	if !IsExist(path) {
		j, _ := json.MarshalIndent(NewEmptyPluginsData(), "", "    ")
		ioutil.WriteFile(path, []byte(j), 0644)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var p PluginsData
	err = json.Unmarshal(b, &p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func ExecPlugin(proc string, args []string) string {
	c := make(chan string)
	go func() {
		out, err := exec.Command(proc, args...).Output()
		if err != nil {
			c <- err.Error()
		} else {
			c <- string(out)
		}
	}()
	return <-c
}
