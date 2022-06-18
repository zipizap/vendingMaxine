package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// Returns []string of rel-path-and-filenames where first is oldest-modified-file and last is newest-modified-file
// filename_regexp filters filenames (without path)
// if dirpath is not readable, it returns empty slice []string
//
// EXAMPLE:
//  fls, err := ioLatestFiles("a/b/c")
//  fls, err := ioLatestFiles("a/b/c","unit-service.flow.*.yaml")
func ioLatestFiles(dirpath string, optional_filename_regexp ...string) ([]string, error) {
	var files []string
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		// dirpath does not exist, return empty slice
		empty_slice := []string{}
		return empty_slice, nil
	}
	fileInfo, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, err
	}
	sort.Slice(fileInfo, func(i, j int) bool {
		t1 := fileInfo[i].ModTime()
		t2 := fileInfo[j].ModTime()
		return t1.Before(t2)
	})
	for _, file := range fileInfo {
		// skip if it's a dir
		if file.IsDir() {
			continue
		}
		// skip if file.Name() does not match filename_regexp
		skip_this_file := false
		for _, a_filename_regexp := range optional_filename_regexp {
			matchTrue, err := regexp.MatchString(a_filename_regexp, file.Name())
			if err != nil {
				return nil, err
			}
			if !matchTrue {
				skip_this_file = true
			}
		}
		if skip_this_file {
			continue
		}
		files = append(files, filepath.Join(dirpath, file.Name()))
	}
	return files, nil
}

func ioUsfDirpath(uname string) (usf_dirpath string) {
	usf_dirpath = filepath.Join("vd-internal", "unit-services", uname)
	return usf_dirpath
}

// "" or filename-relative-path
func ioLatestUsfFile(uname string) (usf_filepath string, err error) {
	usf_dirpath := ioUsfDirpath(uname)
	all_usf_filepaths, err := ioLatestFiles(usf_dirpath, "unit-service.flow.*.yaml")
	if err != nil {
		return "", err
	}
	if len(all_usf_filepaths) == 0 {
		return "", nil
	}
	latest_usf_filepath := all_usf_filepaths[len(all_usf_filepaths)-1]
	return latest_usf_filepath, nil
}

type Dispatcher struct{}

func (v *Dispatcher) GetUserviceFlow(uname string) (usf *UserviceFlow, err error) {
	usf_latest_filepath, err := ioLatestUsfFile(uname)
	if err != nil {
		return nil, err
	}
	if usf_latest_filepath == "" {
		return nil, fmt.Errorf("Could not find any file for UserviceFlow %s\n", uname)
	}
	usf_bytes, err := os.ReadFile(usf_latest_filepath)
	if err != nil {
		return nil, err
	}
	usf = &UserviceFlow{}
	err = yaml.Unmarshal(usf_bytes, usf)
	if err != nil {
		return nil, err
	}
	return usf, nil
}

func (v *Dispatcher) NewUserviceFlow(uname string) (usf *UserviceFlow, err error) {
	var last_usf *UserviceFlow
	// validate that last UserviceFlow (from file!) (if exists), has LatestUpdateStatus "Error|Completed"
	last_usf_filepath, err := ioLatestUsfFile(uname)
	if err != nil {
		return nil, err
	}
	if last_usf_filepath != "" {
		// last_usf_filepath exists
		last_usf, err = v.GetUserviceFlow(uname)
		if err != nil {
			return nil, err
		}
		matchFound, err := regexp.MatchString("Completed|Error", last_usf.Status.Overall.LatestUpdateStatus)
		if err != nil {
			return nil, err
		}
		if !matchFound {
			// a last NewUserviceFlow was found but its neither Complete|Error, so we cannot create a new NewUserviceFlow
			return nil, fmt.Errorf("Cannot create a new UserviceFlow for uname '%s', because there is already an existing UserviceFlow '%s' with Status.Overall.LatestUpdateStatus '%s' (!= Completed|Error)\n", uname, last_usf_filepath, last_usf.Status.Overall.LatestUpdateStatus)
		}
	}

	// Create new usf struct, with uname
	time_now := time.Now()
	usf_filepath := filepath.Join(
		ioUsfDirpath(uname),
		"unit-service.flow."+time_now.Format("20060102150405.00")+".yaml")
	usf = &UserviceFlow{
		Kind: "UnitServiceStatusFlow",
		Name: uname,
	}
	usf.Status.Overall = StatusBlock{
		Name:               uname,
		StartTime:          time_now,
		LatestUpdateTime:   time_now,
		LatestUpdateStatus: "Ongoing_and_locked",
	}

	// Sync usf struct to file
	usf_bytes, err := yaml.Marshal(usf)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(usf_filepath, usf_bytes, 0600)
	if err != nil {
		return nil, err
	}
	return usf, nil
}

func main() {
	d := Dispatcher{}
	uname := "alpha"
	usf, err := d.GetUserviceFlow(uname)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v, %v \n", usf.Status.Overall.LatestUpdateTime, usf.Status.Overall.LatestUpdateStatus)
	usf2, err := d.NewUserviceFlow(uname)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v, %v \n", usf2.Status.Overall.LatestUpdateTime, usf2.Status.Overall.LatestUpdateStatus)
}
