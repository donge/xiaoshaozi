package main

import (
	//"bufio"
	"encoding/json"
	"fmt"
	//"io"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Id        uint
	Type      uint
	Port      string
	Email     string
	Password  string
	AccessKey string
	SecurtKey string
	Quick     uint
	Slow      uint
	QuickInit float64
	SlowInit  float64
	Delta     float64
	Diff      float64
	Pulse     uint
	Clear     bool
	Simulator bool
	Cash      float64
	Coin      float64
}

const CONFIG_FILE = "config.json"

func SaveConfig(file_name string, config *Config) (err error) {
	if file_name == "" {
		file_name = CONFIG_FILE
	}
	fout, err := os.Create(file_name)
	defer fout.Close()
	if err != nil {
		fmt.Println(fout, err)
		return
	}
	/* pretty print */
	b, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Println(err)
		return
	}

	//fout.WriteString("Just a config!\r\n")
	fout.Write(b)
	log.Println("CONFIG SAVED.")
	return
}

func LoadConfig(file_name string, config *Config) (err error) {
	if file_name == "" {
		file_name = CONFIG_FILE
	}
	file, err := os.Open(file_name) // For read access.
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	//r := bufio.NewReader(file)
	//meta_json := r.ReadBytes(io.EOF)
	meta_json, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(meta_json, config)
	if err != nil {
		log.Panic(err)
	}
	/*ra, _ := ioutil.ReadFile("C:\\Windows\\win.ini")*/
	//fmt.Println(config)
	log.Println("CONFIG LOADED.")

	return
}
