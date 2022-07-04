package collection

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

type StatusBlock struct {
	Name                   string                 `yaml:"Name"`
	StartTime              time.Time              `yaml:"StartTime"`
	LatestUpdateTime       time.Time              `yaml:"LatestUpdateTime"`
	LatestUpdateStatus     string                 `yaml:"LatestUpdateStatus"`
	LatestUpdateStatusInfo string                 `yaml:"LatestUpdateStatusInfo"`
	LatestUpdateUml        string                 `yaml:"LatestUpdateUml"`
	LatestUpdateData       map[string]interface{} `yaml:"LatestUpdateData"`
}

// "encoding/json"
type RequestStatusFlow struct {
	Kind       string `yaml:"Kind"`
	Collection string `yaml:"Collection"` // Collection is always defined (since object creation)
	Status     struct {
		Overall           StatusBlock   `yaml:"Overall"`
		ProcessingEngines []StatusBlock `yaml:"ProcessingEngines"`
	} `yaml:"Status"`
}

// if new rsf can be created => yes == true
func (rsf *RequestStatusFlow) checkNewRsfCanBeCreated() (yes bool, err error) {
	// NewRsf can be created if
	//    a) LastRst does not exist
	//  or
	//    b) LastRst has .Status.Overall.LatestUpdateStatus ~= Error|Completed

	// Lets check a)
	col, err := rsf.i1_getCollection()
	if err != nil {
		return false, err
	}
	lastFiles, err := col.lastFiles()
	if err != nil {
		return false, err
	}
	if len(lastFiles) == 0 {
		// LastRst does not exist
		return true, nil
	}

	// Lets check b)
	lastRsf, err := col.LastRsf()
	if err != nil {
		return false, err
	}
	matchFound, err := regexp.MatchString("Completed|Error", lastRsf.Status.Overall.LatestUpdateStatus)
	if err != nil {
		return false, err
	}
	if matchFound {
		return true, nil
	} else {
		return false, nil
	}
}

// // Validates if last-userviceflow-of-uname is Error|Completed (or does not exist) and if so then
// // creates a new userviceflow file (unit-service.flow.xxxx.yaml) setting its
// // .Kind|Name, .Status.Overall.Name|StartTime|LatestUpdateTime|LatestUpdateStatus (Ongoing_and_locked)
// func (v *Dispatcher) NewUserviceFlow(uname string) (cantBeDone bool, usf *UserviceFlow, err error) {
// 	var last_usf *UserviceFlow
// 	// validate that last UserviceFlow (from file!) (if exists), has LatestUpdateStatus "Error|Completed"
// 	last_usf_filepath, err := ioLatestUsfFile(uname)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if last_usf_filepath != "" {
// 		// last_usf_filepath exists
// 		last_usf, err = v.GetUserviceFlow(uname)
// 		if err != nil {
// 			return nil, err
// 		}
// 		matchFound, err := regexp.MatchString("Completed|Error", last_usf.Status.Overall.LatestUpdateStatus)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if !matchFound {
// 			// a last NewUserviceFlow was found but its neither Complete|Error, so we cannot create a new NewUserviceFlow
// 			return nil, fmt.Errorf("cannot create a new UserviceFlow for uname '%s', because there is already an existing UserviceFlow '%s' with Status.Overall.LatestUpdateStatus '%s' (!= Completed|Error)", uname, last_usf_filepath, last_usf.Status.Overall.LatestUpdateStatus)
// 		}
// 	}
// 	// Create new usf struct, with uname
// 	time_now := time.Now()
// 	usf = &UserviceFlow{
// 		Kind: "UnitServiceStatusFlow",
// 		Name: uname,
// 	}
// 	usf.Status.Overall = StatusBlock{
// 		Name:               uname,
// 		StartTime:          time_now,
// 		LatestUpdateTime:   time_now,
// 		LatestUpdateStatus: "Ongoing_and_locked",
// 	}
// 	// Sync usf struct to file
// 	err = ioSyncUsf2File(usf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return usf, nil
// }

// At start of this function, its expected:
//   - rsf has all empty-values, except rsf.Collection that is filled-up
//		.Collection
//
// This function will then:
//    - rsf.checkNewRsfCanBeCreated()
//	  - fill in the rsf struct with data from web_sblock, and syncSave to yaml file
//    - call rsf.i1_runProcessingEngines() which will be launched (async) and left running asynchronously
//
func (rsf *RequestStatusFlow) new_from_webConsumerSelection(web_sblock StatusBlock) (cantDo bool, err error) {
	//    - rsf.checkNewRsfCanBeCreated()
	yes, err := rsf.checkNewRsfCanBeCreated()
	if err != nil {
		return true, err
	}
	if !yes {
		// new rsf cannot be created => cantDo
		return true, nil
	}

	//	  - fill in the rsf struct with data from web_sblock, and syncSave to yaml file
	rsf.Kind = "RequestStatusFlow"
	c, err := rsf.i1_getCollection()
	if err != nil {
		return true, err
	}
	rsf.Status.Overall = StatusBlock{
		Name:               c.Name,
		StartTime:          web_sblock.LatestUpdateTime,
		LatestUpdateTime:   web_sblock.LatestUpdateTime,
		LatestUpdateStatus: "Ongoing_and_locked",
		LatestUpdateData: map[string]interface{}{
			"consumer-selection.previous.json": web_sblock.LatestUpdateData["consumer-selection.previous.json"],
			"consumer-selection.next.json":     web_sblock.LatestUpdateData["consumer-selection.next.json"],
		},
	}
	rsf.Status.ProcessingEngines = append(rsf.Status.ProcessingEngines, web_sblock)

	// create rsf-yaml-file, empty for now (after rsf.checkNewRsfCanBeCreated())
	// NOTE: this file is now empty but will latter be picked-up by rsf.i2_syncSaveToLastFile
	// which will then write into the file the rsf struct contents :)
	new_rsf_filepath := filepath.Join(
		c.dirpath(),
		"RequestStatusFlow."+rsf.Status.Overall.StartTime.Format("20060102150405.00")+".yaml")
	file, err := os.Create(new_rsf_filepath)
	if err != nil {
		return true, err
	}
	file.Close()
	// at this point, the lastfile is the empty-file we just created. So with i2_syncSaveToLastFile() we will actually
	// write the rsf struct into the file :)
	rsf.i2_syncSaveToLastFile()

	//    - call rsf.i1_runProcessingEngines() which will be launched (async) and left running asynchronously
	go rsf.i1_runProcessingEngines()
	return false, nil
}

// This function should always be called into a go-routine: "go runProcessingEngines()"
func (rsf *RequestStatusFlow) i1_runProcessingEngines() {

	// for _, a_ProcEngine_Binary := range procenginebinaries {
	// 	rsf.i1_append_Sblock(...)
	// 	//execute a_ProcEngine_Binary and get stdouterr_bytes + exitcode
	// 	rsf.i1_update_LastSblock(...)
	//   }

	procEng_binaries, err := rsf.i1_getProcEngineBinariesList()
	if err != nil {
		// todo log error
		return
	}

	procEng_binaries_indexLastElement := len(procEng_binaries) - 1
	var i_sblock_is_the_finalSblock bool
	for i, i_procEng_binary_filepath := range procEng_binaries {

		//
		// a) In the start:
		//   - a.1) Append i_sblock "Ongoing_and_locked" before executing the engine
		//			LatestUpdateStatus:  "Ongoing_and_locked"
		//			LatestUpdateStatusInfo: "Running"
		//			LatestUpdateData:
		//		  		"consumer-selection.previous.json": <gz.b64> = Overall.LatestUpdateData["consumer-selection.previous.json"]
		//					>> added by dispatcher at procEng-start
		//		  		"consumer-selection.next.json": <gz.b64>     = Overall.LatestUpdateData["consumer-selection.next.json"]
		//					>> added at procEng-start
		//
		// b) Execute i_procEng_binary_filepath and get stdouterr_bytes + exitcode into i_sblock
		//	  - b.0) selectionPreviousJson_bytes, selectionNextJson_bytes
		// 		  selectionPreviousJson_bytes 	is read from Overall.LatestUpdateData["consumer-selection.previous.json"]
		// 		  selectionNextJson_bytes 		is read from Overall.LatestUpdateData["consumer-selection.next.json"]
		//	  - b.1) selectionPreviousJson_bytes	is saved into 	i_procEng_selectionPreviousJson_filepath
		//	  - b.2) selectionNextJson_bytes 		is saved into 	i_procEng_selectionNextJson_filepath
		//	  - b.3) i_procEng_binary_filepath 	is executed and we get stdouterr_bytes + exitcode
		//	  - b.4) selectionNextJson_bytes 		is re-read from i_procEng_selectionNextJson_filepath (file might have been modified by engine)
		//
		// c) In the end:
		//	  - c.1) Update i_sblock:
		//			LatestUpdateStatus:    Error ($? > 0) | Completed ($?==0)
		//			LatestUpdateStatusInfo: (stdout+stderr)
		//			LatestUpdateData:
		//		  		"consumer-selection.next.json": <gz.b64>   (overwritten with selectionNextJson_bytes)
		//
		//	  - c.2) Update Overall:
		//			LatestUpdateData:
		//		  		"consumer-selection.next.json": <gz.b64>   (overwritten with selectionNextJson_bytes)
		//
		//

		// a) In the start:
		//   - a.1) Append i_sblock "Ongoing_and_locked" before executing the engine
		//			LatestUpdateStatus:  "Ongoing_and_locked"
		//			LatestUpdateStatusInfo: "Running"
		//			LatestUpdateData:
		//		  		"consumer-selection.previous.json": <gz.b64> = Overall.LatestUpdateData["consumer-selection.previous.json"]
		//					>> added by dispatcher at procEng-start
		//		  		"consumer-selection.next.json": <gz.b64>     = Overall.LatestUpdateData["consumer-selection.next.json"]
		//					>> added at procEng-start
		t_now := time.Now()
		i_sblock := StatusBlock{
			Name:                   filepath.Base(i_procEng_binary_filepath),
			StartTime:              t_now,
			LatestUpdateTime:       t_now,
			LatestUpdateStatus:     "Ongoing_and_locked",
			LatestUpdateStatusInfo: "Running",
		}
		i_sblock.LatestUpdateData["consumer-selection.previous.json"] = rsf.Status.Overall.LatestUpdateData["consumer-selection.previous.json"]
		i_sblock_is_the_finalSblock = (i == procEng_binaries_indexLastElement)
		cantDo, err := rsf.i1_append_Sblock(i_sblock_is_the_finalSblock, i_sblock)
		if err != nil {
			// todo: log error
			return
		}
		if cantDo {
			// todo: log error
			return
		}
		// ATP: i_sblock was appended successfully

		// b) Execute i_procEng_binary_filepath and get stdouterr_bytes + exitcode into i_sblock
		//	  - b.0) selectionPreviousJson_bytes, selectionNextJson_bytes
		// 		  selectionPreviousJson_bytes 	is read from Overall.LatestUpdateData["consumer-selection.previous.json"]
		// 		  selectionNextJson_bytes 		is read from Overall.LatestUpdateData["consumer-selection.next.json"]
		selectionPreviousJson_bytes, err := rsf.i1_decode_gzB64_to_bytes(rsf.Status.Overall.LatestUpdateData["consumer-selection.previous.json"].(string))
		if err != nil {
			// todo log error
			return
		}
		selectionNextJson_bytes, err := rsf.i1_decode_gzB64_to_bytes(rsf.Status.Overall.LatestUpdateData["consumer-selection.next.json"].(string))
		if err != nil {
			// todo log error
			return
		}

		//	  - b.1) selectionPreviousJson_bytes	is saved into 	i_procEng_selectionPreviousJson_filepath
		//	  - b.2) selectionNextJson_bytes 		is saved into 	i_procEng_selectionNextJson_filepath
		i_procEng_tmpSubdirpath, err := ioutil.TempDir("", "tmpsubdir")
		if err != nil {
			// todo log error
			return
		}
		defer os.RemoveAll(i_procEng_tmpSubdirpath)
		i_procEng_selectionPreviousJson_filepath := filepath.Join(i_procEng_tmpSubdirpath, "consumer-selection.previous.json")
		i_procEng_selectionNextJson_filepath := filepath.Join(i_procEng_tmpSubdirpath, "consumer-selection.next.json")
		err = os.WriteFile(i_procEng_selectionPreviousJson_filepath, selectionPreviousJson_bytes, 0644)
		if err != nil {
			// todo log error
			return
		}
		err = os.WriteFile(i_procEng_selectionNextJson_filepath, selectionNextJson_bytes, 0644)
		if err != nil {
			// todo log error
			return
		}
		//	  - b.3) i_procEng_binary_filepath 	is executed and we get stdouterr_bytes + exitcode
		var exitcode int
		cmd := exec.Command(i_procEng_binary_filepath, i_procEng_selectionPreviousJson_filepath, i_procEng_selectionNextJson_filepath)
		stdouterr_bytes, err := cmd.CombinedOutput()
		if err != nil {
			// err happened, either because of exitCode != 0 or because of another internal failure
			if exitError, ok := err.(*exec.ExitError); ok {
				// err happened because exitCode != 0
				exitcode = exitError.ExitCode()
			} else {
				// err happened because of another internal failure
				// todo: log err
				// todo: this should propagate to sblock "Error"
				return
			}
		} else {
			// err didn't happened, so exitcode = 0
			exitcode = 0
		}
		// ATP: exitcode, stdouterr_bytes are ready

		//	  - b.4) selectionNextJson_bytes 		is re-read from i_procEng_selectionNextJson_filepath (file might have been modified by engine)
		selectionNextJson_bytes, err = os.ReadFile(i_procEng_selectionNextJson_filepath)
		if err != nil {
			// todo: log error
			return
		}

		// c) In the end:
		//	  - c.1) Update i_sblock:
		//			LatestUpdateStatus:    Error ($? > 0) | Completed ($?==0)
		//			LatestUpdateStatusInfo: (stdout+stderr)
		//			LatestUpdateData:
		//		  		"consumer-selection.next.json": <gz.b64>   (overwritten with selectionNextJson_bytes
		i_sblock.LatestUpdateTime = time.Now()
		if exitcode == 0 {
			i_sblock.LatestUpdateStatus = "Completed"
		} else {
			// exitcode != 0
			i_sblock.LatestUpdateStatus = "Error"
		}

		i_sblock.LatestUpdateStatusInfo = string(stdouterr_bytes)
		selectionNextJson_gzB64, err := rsf.i1_encode_bytes_to_gzB64(selectionNextJson_bytes)
		if err != nil {
			// todo: log error
			return
		}
		i_sblock.LatestUpdateData["consumer-selection.next.json"] = selectionNextJson_gzB64

		cantDo, err = rsf.i1_update_LastSblock(i_sblock_is_the_finalSblock, i_sblock)
		if err != nil {
			// todo: log error
			return
		}
		if cantDo {
			// todo: log error
			return
		}
		// ATP: i_sblock was updated successfully

		//	  - c.2) Update Overall:
		//			LatestUpdateData:
		//		  		"consumer-selection.next.json": <gz.b64>   (overwritten with selectionNextJson_bytes)
		new_LatestUpdateData := make(map[string]interface{})
		new_LatestUpdateData["consumer-selection.next.json"] = selectionNextJson_gzB64
		cantDo, err = rsf.i1_update_Overall_LatestUpdateData(new_LatestUpdateData)
		if err != nil {
			// todo: log error
			return
		}
		if cantDo {
			// todo: log error
			return
		}

	} // end for
} // end func

// Reset rsf and loads it again from file contents
func (rsf *RequestStatusFlow) i1_syncLoadFromLastFile() (err error) {
	col, err := rsf.i1_getCollection()
	if err != nil {
		return err
	}
	lastFiles, err := col.lastFiles()
	if err != nil {
		return err
	}
	lastFile_filepath := lastFiles[0]
	lastFile_bytes, err := os.ReadFile(lastFile_filepath)
	if err != nil {
		return err
	}
	// reset rsf before loading it from file
	empty_rsf := &RequestStatusFlow{}
	*rsf = *empty_rsf
	err = yaml.Unmarshal(lastFile_bytes, rsf)
	if err != nil {
		return err
	}
	return nil
}

func (rsf *RequestStatusFlow) i1_getCollection() (col *Collection, err error) {
	col, err = GetCollection(rsf.Collection)
	if err != nil {
		return nil, err
	}
	return col, err
}

// It will syncLoad rsf from file,
// append new_sblock into rsf.Status.ProcessingEngines and recalculate rsf.Status.Overall.LatestUpdateStatus [1],
// and syncSave rsf into file
//
// [1] See diagram, scheme titled "LatestUpdateStatus Transitions - Overall & ProcessingEngines"
func (rsf *RequestStatusFlow) i1_append_Sblock(new_sblock_is_finalSblock bool, new_sblock StatusBlock) (cantDo bool, err error) {
	cantDo, err = rsf.i2_append_or_update_Sblock("append", new_sblock_is_finalSblock, new_sblock)
	if err != nil {
		return true, err
	}
	return cantDo, nil
}

// It will syncLoad rsf from file,
// update rsf.Status.ProcessingEngines[-1] with new_sblock and recalculate rsf.Status.Overall.LatestUpdateStatus [1],
// and syncSave rsf into file
//
// [1] See diagram, scheme titled "LatestUpdateStatus Transitions - Overall & ProcessingEngines"
func (rsf *RequestStatusFlow) i1_update_LastSblock(new_sblock_is_finalSblock bool, new_sblock StatusBlock) (cantDo bool, err error) {
	cantDo, err = rsf.i2_append_or_update_Sblock("update", new_sblock_is_finalSblock, new_sblock)
	if err != nil {
		return true, err
	}
	return cantDo, nil
}

// This function will merge key/values in new_LatestUpdateData into rsf.Status.Overall.LatestUpdataData
// ie, it will overwrite rsf.Status.Overall.LatestUpdataData with the key/values found new_LatestUpdateData
//
// It will syncLoad rsf from file,
// update rsf.Status.Overall.LatestUpdateData so it gets merged with new_LatestUpdateData
// and syncSave rsf into file
//
// Example usage:
//
//   var new_LatestUpdateData map[string]interface{}
//   new_LatestUpdateData["consumer-selection.next.json"] = selectionNextJson_gzB64
//   cantDo, err = rsf.i1_update_Overall_LatestUpdateData(new_LatestUpdateData)
//   if err != nil {
//   	 return ...
//   }
//   if cantDo == true {
//   	 return ...
//   }
func (rsf *RequestStatusFlow) i1_update_Overall_LatestUpdateData(new_LatestUpdateData map[string]interface{}) (cantDo bool, err error) {
	// Assure rsf struct is sync'ed from rsf-file
	err = rsf.i1_syncLoadFromLastFile()
	if err != nil {
		return true, err
	}

	// Validate rsf can be updated
	yes, err := rsf.i2_checkRsfCanBeUpdated()
	if err != nil {
		return true, err
	}
	if !yes {
		// rsf cannot be updated (either to _append_Sblock or _update_Sblock)
		return true, nil
	}
	// rsf can now be updated

	// Update rsf.Status.Overall.LatestUpdateData with new_LatestUpdateData
	for k, v := range new_LatestUpdateData {
		rsf.Status.Overall.LatestUpdateData[k] = v
	}

	// Save rsf struct by overwritting it into rsf-file
	err = rsf.i2_syncSaveToLastFile()
	if err != nil {
		return true, err
	}
	return false, nil
}

func (rsf *RequestStatusFlow) i1_encode_bytes_to_gzB64(bytes []byte) (gzB64 string, err error) {
	gz, err := gZipData(bytes)
	if err != nil {
		return "", err
	}
	gzB64 = base64.StdEncoding.EncodeToString(gz)
	return gzB64, nil
}

func (rsf *RequestStatusFlow) i1_decode_gzB64_to_bytes(gzB64 string) (bytes []byte, err error) {
	gz, err := base64.StdEncoding.DecodeString(gzB64)
	if err != nil {
		return nil, err
	}
	bytes, err = gUnzipData(gz)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

//  when rfs can be updated: yes == true
func (rsf *RequestStatusFlow) i2_checkRsfCanBeUpdated() (yes bool, err error) {
	// rsf can be updated if has .Status.Overall.LatestUpdateStatus != Error|Completed
	matchFound, err := regexp.MatchString("Completed|Error", rsf.Status.Overall.LatestUpdateStatus)
	if err != nil {
		return false, err
	}
	if matchFound {
		// Error|Completed => cannot be updated
		return false, nil
	} else {
		// can be updated
		return true, nil
	}
}

// Saves rsf into lastFile
func (rsf *RequestStatusFlow) i2_syncSaveToLastFile() (err error) {
	col, err := rsf.i1_getCollection()
	if err != nil {
		return err
	}
	lastFiles, err := col.lastFiles()
	if err != nil {
		return err
	}
	lastFile_filepath := lastFiles[0]

	lastFile_bytes, err := yaml.Marshal(rsf)
	if err != nil {
		return err
	}
	err = os.WriteFile(lastFile_filepath, lastFile_bytes, 0600)
	if err != nil {
		return err
	}
	return nil
}

// append_or_update == "append" or "update"
func (rsf *RequestStatusFlow) i2_append_or_update_Sblock(append_or_update string, new_sblock_is_finalSblock bool, new_sblock StatusBlock) (cantDo bool, err error) {
	// Validate if new_sblock is correct
	// a) new_sblock.LatestUpdateStatus ~= Ongoing_and_locked|Completed|Error
	a_match_was_found := regexp.MustCompile(`Ongoing_and_locked|Completed|Error`).MatchString(new_sblock.LatestUpdateStatus)
	if !a_match_was_found {
		// there is no match, this new_sblock.LatestUpdateStatus is unknown
		return true, fmt.Errorf("invalid new_sblock.LatestUpdateStatus '%s', was expecting Ongoing_and_locked|Completed|Error", new_sblock.LatestUpdateStatus)
	}

	// Assure rsf struct is sync'ed from rsf-file
	err = rsf.i1_syncLoadFromLastFile()
	if err != nil {
		return true, err
	}

	// Validate rsf can be updated with new_sblock (either to _append_Sblock or _update_Sblock)
	yes, err := rsf.i2_checkRsfCanBeUpdated()
	if err != nil {
		return true, err
	}
	if !yes {
		// rsf cannot be updated (either to _append_Sblock or _update_Sblock)
		return true, nil
	}
	// rsf can now be updated, lets append-or-update new_sblock

	if append_or_update == "append" {
		// Append new_sblock into rsf struct
		rsf.Status.ProcessingEngines = append(rsf.Status.ProcessingEngines, new_sblock)
	} else if append_or_update == "update" {
		// Update rsf struct, to update LastSblock to be new_sblock
		rsf.Status.ProcessingEngines[len(rsf.Status.ProcessingEngines)-1] = new_sblock
	} else {
		return true, fmt.Errorf("unexpected argument append_or_update '%s'", append_or_update)
	}

	// Now that new_sblock was appended as last-sblock of rsf struct, lets recalculate OverallStatus of rsf struct
	err = rsf.i2_autoupdate_OverallStatus(new_sblock_is_finalSblock)
	if err != nil {
		return true, err
	}

	// Save rsf struct by overwritting it into rsf-file
	err = rsf.i2_syncSaveToLastFile()
	if err != nil {
		return true, err
	}
	return false, nil
}

// This function reads lastSblock.LatestUpdateStatus and updates Overall.LatestUpdateStatus accordingly (or leaves unchanged)
// See diagram, scheme titled "LatestUpdateStatus Transitions - Overall & ProcessingEngines"
//
// This functiond does not check if Overall.LatestUpdateStatus is valid-to-be-updated or not as it expects the calling
// function to already have made that validation before calling this.
// It assumes the Overall.LatestUpdateStatus is correct (ex:like "Ongoing_and_locked") and will not modify it, unless:
//     LastSBlock.LatestUpdateStatus ~= "Error"
// in which case it will overwrite
//     Overall.LatestUpdateStatus = "Error"
//
// :) Think very well before trying to change Status and transitions, as it implies changes in this function and
// many other functions besides this one :)
func (rsf *RequestStatusFlow) i2_autoupdate_OverallStatus(lastSblock_is_finalSblock bool) (err error) {

	lastSb := rsf.Status.ProcessingEngines[len(rsf.Status.ProcessingEngines)-1]
	lastSbStatus := lastSb.LatestUpdateStatus
	if !lastSblock_is_finalSblock {
		// lastSblock is not the finalSblock, its a mid-chain Sblock
		switch {
		case regexp.MustCompile(`Ongoing_and_locked|Completed`).MatchString(lastSbStatus):
			// lastSbStatus is "Ongoing_and_locked" or "Completed"
			// so Overall.LatestUpdateStatus is not modified (left unchanged)
			// Do nothing
		case regexp.MustCompile(`Error`).MatchString(lastSbStatus):
			// lastSbStatus is "Error"
			// so Overall.LatestUpdateStatus is updated (overwritten):
			//    Overall.LatestUpdateStatus = "Error"
			rsf.Status.Overall.LatestUpdateStatus = "Error"
		default:
			// lastSbStatus is not recognized - error out
			return fmt.Errorf("!crash boom bang! unrecognized lastSbStatus '%s'", lastSbStatus)
		}

	} else {
		// lastSblock is the finalSblock
		switch {
		case regexp.MustCompile(`Ongoing_and_locked`).MatchString(lastSbStatus):
			// lastSbStatus is "Ongoing_and_locked"
			// so Overall.LatestUpdateStatus is not modified (left unchanged)
			// Do nothing
		case regexp.MustCompile(`Completed`).MatchString(lastSbStatus):
			// lastSbStatus is "Completed"
			// so Overall.LatestUpdateStatus is updated (overwritten):
			//    Overall.LatestUpdateStatus = "Completed"
			rsf.Status.Overall.LatestUpdateStatus = "Completed"
		case regexp.MustCompile(`Error`).MatchString(lastSbStatus):
			// lastSbStatus is "Error"
			// so Overall.LatestUpdateStatus is updated (overwritten):
			//    Overall.LatestUpdateStatus = "Error"
			rsf.Status.Overall.LatestUpdateStatus = "Error"
		default:
			// lastSbStatus is not recognized - error out
			return fmt.Errorf("!crash boom bang! unrecognized lastSbStatus '%s' - aborting", lastSbStatus)
		}
	}
	return nil
}

// Returns processing-engine-binary files, sorted alfabetically, in procEngineBinariesFilepaths.
//
// The matched files are "config/processingEngines/*.engine" and must be executable.
// Subdirs and non-matching files are simply ignored.
func (rsf *RequestStatusFlow) i1_getProcEngineBinariesList() (procEngineBinariesFilepaths []string, err error) {
	optional_filename_regexp := []string{`\.engine$`}
	dirpath := filepath.Join("config", "processingEngines")
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		// dirpath does not exist, return empty slice
		procEngineBinariesFilepaths = []string{}
		return procEngineBinariesFilepaths, nil
	}
	fileInfo, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, err
	}
	sort.Slice(fileInfo, func(i, j int) bool {
		return fileInfo[i].Name() < fileInfo[j].Name()
	})
	for _, file := range fileInfo {
		if file.IsDir() {
			// skip if it's a dir
			continue
		} else if file.Mode()&0111 == 0 {
			// skip if file is not executable (+x)
			continue
		}
		if len(optional_filename_regexp) > 0 {
			// skip if file.Name() does not match any filename_regexp
			include_this_file := false
			for _, a_filename_regexp := range optional_filename_regexp {
				matchTrue, err := regexp.MatchString(a_filename_regexp, file.Name())
				if err != nil {
					return nil, err
				}
				if matchTrue {
					include_this_file = true
				}
			}
			if !include_this_file {
				continue
			}
		}
		procEngineBinariesFilepaths = append(procEngineBinariesFilepaths, filepath.Join(dirpath, file.Name()))
	}
	return procEngineBinariesFilepaths, nil
}

// https://gist.github.com/alex-ant/aeaaf497055590dacba760af24839b8d
func gUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}

	resData = resB.Bytes()

	return
}

func gZipData(data []byte) (compressedData []byte, err error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	_, err = gz.Write(data)
	if err != nil {
		return
	}

	if err = gz.Flush(); err != nil {
		return
	}

	if err = gz.Close(); err != nil {
		return
	}

	compressedData = b.Bytes()

	return
}
