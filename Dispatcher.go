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
	all_usf_filepaths, err := ioLatestFiles(usf_dirpath, "^unit-service.flow.*.yaml$")
	if err != nil {
		return "", err
	}
	if len(all_usf_filepaths) == 0 {
		return "", nil
	}
	latest_usf_filepath := all_usf_filepaths[len(all_usf_filepaths)-1]
	return latest_usf_filepath, nil
}

// Syncs usf into corresponding file (creates or overwrites file)
// Expects usf .Name and .Status.Overall.StartTime already defined, will be used to infer usf_filepath
func ioSyncUsf2File(usf *UserviceFlow) (err error) {
	// get usf_filepath
	usf_filepath := filepath.Join(
		ioUsfDirpath(usf.Name),
		"unit-service.flow."+usf.Status.Overall.StartTime.Format("20060102150405.00")+".yaml")
	if err != nil {
		return err
	}
	// save usf into usf_filepath
	usf_bytes, err := yaml.Marshal(usf)
	if err != nil {
		return err
	}
	err = os.WriteFile(usf_filepath, usf_bytes, 0600)
	if err != nil {
		return err
	}
	return nil
}

// Reads lastSblock from usf.Status.ProcessingEngines, and updates usf.Status.Overall
func usfAutoUpdStatusOverall(usf *UserviceFlow) (cannotUpdate bool, err error) {
	// Check: usf cannot be updated when .Status.Overall.LatestUpdateStatus = "Completed|Error"
	matchOk, err := regexp.MatchString("^Completed|Error", usf.Status.Overall.LatestUpdateStatus)
	if err != nil {
		return err
	}
	if !matchOk {
		// There is no match => usf is not Completed|Error => cannot update it
		return fmt.Errorf()
	}

	lastSblock := usf.Status.ProcessingEngines[len(usf.Status.ProcessingEngines)-1]

	// **********************************************************************************************************************
	//   lastSblock                        ->          .Status.Overall
	//     .LatestUpdateStatus                           .LatestUpdateStatus                  .LatestUpdateStatusInfo
	//        Ongoing_and_locked           ->              Ongoing_and_locked (no change)        "Running xxxxx"
	// 	      Completed                    ->              Ongoing_and_locked (no change)        "Finished running xxxxx"
	// 	      Error                        ->              Error                                 "Error running xxxxxx"
	//
	// **********************************************************************************************************************

	// Update Status.Overall
	usf.Status.Overall.LatestUpdateTime = lastSblock.LatestUpdateTime

	switch newSblock_LatestUpdateStatus := lastSblock.LatestUpdateStatus; newSblock_LatestUpdateStatus {
	case "Ongoing_and_locked":
		usf.Status.Overall.LatestUpdateStatus = "Ongoing_and_locked"
		usf.Status.Overall.LatestUpdateStatusInfo = "Running " + lastSblock.Name
	case "Completed":
		usf.Status.Overall.LatestUpdateStatus = "Ongoing_and_locked"
		usf.Status.Overall.LatestUpdateStatusInfo = "Finished running " + lastSblock.Name
	case "Error":
		usf.Status.Overall.LatestUpdateStatus = "Error"
		usf.Status.Overall.LatestUpdateStatusInfo = "Error running " + lastSblock.Name
	default:
		return fmt.Errorf("lastSblock.LatestUpdateStatus '%s' unknown (expected Completed|Error|Ongoing_and_locked)", lastSblock.LatestUpdateStatus)
	}
	return nil
}

type Dispatcher struct{}

// Reads corresponding unit-service.flow.xxx.yaml file (must exist!) and returns usf pointer
func (v *Dispatcher) GetUserviceFlow(uname string) (usf *UserviceFlow, err error) {
	usf_latest_filepath, err := ioLatestUsfFile(uname)
	if err != nil {
		return nil, err
	}
	if usf_latest_filepath == "" {
		return nil, fmt.Errorf("could not find any file for UserviceFlow %s", uname)
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

// Validates if last-userviceflow-of-uname is Error|Completed (or does not exist) and if so then
// creates a new userviceflow file (unit-service.flow.xxxx.yaml) setting its
// .Kind|Name, .Status.Overall.Name|StartTime|LatestUpdateTime|LatestUpdateStatus (Ongoing_and_locked)
func (v *Dispatcher) NewUserviceFlow(uname string) (cantBeDone bool, usf *UserviceFlow, err error) {
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
			return nil, fmt.Errorf("cannot create a new UserviceFlow for uname '%s', because there is already an existing UserviceFlow '%s' with Status.Overall.LatestUpdateStatus '%s' (!= Completed|Error)", uname, last_usf_filepath, last_usf.Status.Overall.LatestUpdateStatus)
		}
	}

	// Create new usf struct, with uname
	time_now := time.Now()
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
	err = ioSyncUsf2File(usf)
	if err != nil {
		return nil, err
	}
	return usf, nil
}

func (v *Dispatcher) UserviceFlow_Append_Sblock(uname string, sblock *StatusBlock) (usf *UserviceFlow, err error) {
	usf, err = v.GetUserviceFlow(uname)
	if err != nil {
		return nil, err
	}
	// Validate sblock
	// --nothing for now--

	// Append sblock into Status.ProcessingEngines[]
	usf.Status.ProcessingEngines = append(usf.Status.ProcessingEngines, *sblock)

	// Update Status.Overall
	err = usfAutoUpdStatusOverall(usf)
	if err != nil {
		return nil, err
	}

	// Sync usf to file - this should only be done when everything else is calculated and correct (last step)
	err = ioSyncUsf2File(usf)
	if err != nil {
		return nil, err
	}
	return usf, nil
}

func (v *Dispatcher) UserviceFlow_Update_LastSblock(uname string, sblock *StatusBlock) (usf *UserviceFlow, err error) {
	usf, err = v.GetUserviceFlow(uname)
	if err != nil {
		return nil, err
	}
	// Validate sblock
	// --nothing for now--

	// Update Last sblock of Status.ProcessingEngines[]
	slen := len(usf.Status.ProcessingEngines)
	if slen == 0 {
		// Protection when []ProcessingEngines is empty => cannot update lastblock
		return nil, fmt.Errorf("usf of uname '%s' has .Status.ProcessingEngines as an empty slice, so its not possible to update its lastSblock", uname)
	}
	usf.Status.ProcessingEngines[slen-1] = *sblock

	// Update Status.Overall
	err = usfAutoUpdStatusOverall(usf)
	if err != nil {
		return nil, err
	}

	// Sync usf to file - this should only be done when everything else is calculated and correct (last step)
	err = ioSyncUsf2File(usf)
	if err != nil {
		return nil, err
	}
	return usf, nil
}

// Reads last_sblock from usf, and propagates its state into .State.Overall:
//       last_sblock          ->       .State.Overall
//       LatestUpdateStatus             copied
func (v *Dispatcher) UserviceFlow_NoMoreSblocks(uname string) (err error) {
	usf, err = v.GetUserviceFlow(uname)
	if err != nil {
		return nil, err
	}

	// read last_sblock of Status.ProcessingEngines[]
	slen := len(usf.Status.ProcessingEngines)
	if slen == 0 {
		// Protection when []ProcessingEngines is empty => cannot read lastblock
		return fmt.Errorf("usf of uname '%s' has .Status.ProcessingEngines as an empty slice, so its not possible to read its lastSblock (this is unexpected)", uname)
	}
	last_sblock := usf.Status.ProcessingEngines[slen-1]

	// Update Status.Overall
	err = usfAutoUpdStatusOverall(usf)
	if err != nil {
		return nil, err
	}

	// Sync usf to file
	err = ioSyncUsf2File(usf)
	if err != nil {
		return nil, err
	}
	return usf, nil
}

func main() {
	d := Dispatcher{}
	uname := "alpha"
	usf1, err := d.GetUserviceFlow(uname)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v, %v \n", usf1.Status.Overall.LatestUpdateTime, usf1.Status.Overall.LatestUpdateStatus)
	// {
	// 	usf2, err := ds.NewUserviceFlow(uname)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Printf("%v, %v \n", usf2.Status.Overall.LatestUpdateTime, usf2.Status.Overall.LatestUpdateStatus)
	// }

	t := time.Now()
	d.UserviceFlow_Append_Sblock(
		uname,
		&StatusBlock{
			Name:                   "TestBlock0",
			StartTime:              t,
			LatestUpdateTime:       t,
			LatestUpdateStatus:     "Completed",
			LatestUpdateStatusInfo: "I'm completed",
		},
	)
	usf1, err = d.GetUserviceFlow(uname)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v, %v \n", usf1.Status.Overall.LatestUpdateTime, usf1.Status.Overall.LatestUpdateStatus)
}
