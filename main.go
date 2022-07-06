package main

import (
	"fmt"
	"proto-VD/collection"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	spew.Config.Indent = "  "
	//    log.Fatal(err)
	//    log.Info("info message")
}

func main() {
	col, err := collection.GetCollection("alpha")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("rsf_last")
	rsf_last, err := col.LastRsf()
	if err != nil {
		log.Fatal(err)
	}
	spew.Dump(rsf_last)

	// fmt.Println("-------- rsf_new ---------")
	// t_now := time.Now()
	// cantDo, rsf_new, err := col.NewRsf(collection.StatusBlock{
	// 	Name:                   "WebConsumerSelection",
	// 	StartTime:              t_now,
	// 	LatestUpdateTime:       t_now,
	// 	LatestUpdateStatus:     "Completed",
	// 	LatestUpdateStatusInfo: "",
	// 	LatestUpdateUml:        "",
	// 	LatestUpdateData: map[string]interface{}{
	// 		"consumer-selection.previous.json": "H4sIAAAAAAAAA6vmUgACpdK8zBLd4tSisszk1GIlK4VosHA1mAQryEvMTQWKKyXmFGQkKukgJIqLM7JTK8F6YpGEc/Pz0vN1U5IQhqEaiGJoUWqKElymFs5CNq+qSslKKSoKogyiJJarFgA7EgRyvgAAAA==",
	// 		"consumer-selection.next.json":     "H4sIAAAAAAAAA6vmUgACpdK8zBLd4tSisszk1GIlK4VosHA1mAQryEvMTQWKKyXmFGQkKukgJIqLM7JTK8F6YpGEc/Pz0vN1U5IQhqEaiGJoUWqKElymFs5CNq+qSslKKSoKogyiJJarFgA7EgRyvgAAAA==",
	// 	},
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if cantDo {
	// 	log.Fatal("Just cantDo now")
	// }
	//
	fmt.Println("-------- rsf_new via NewRsf_from_WebconsumerSelection ---------")
	webdata := map[string]interface{}{
		"products.schema.json":             "",
		"consumer-selection.previous.json": "H4sIAAAAAAAAA6vmUgACpdK8zBLd4tSisszk1GIlK4VosHA1mAQryEvMTQWKKyXmFGQkKukgJIqLM7JTK8F6YpGEc/Pz0vN1U5IQhqEaiGJoUWqKElymFs5CNq+qSslKKSoKogyiJJarFgA7EgRyvgAAAA==",
		"consumer-selection.next.json":     "H4sIAAAAAAAAA6vmUgACpdK8zBLd4tSisszk1GIlK4VosHA1mAQryEvMTQWKKyXmFGQkKukgJIqLM7JTK8F6YpGEc/Pz0vN1U5IQhqEaiGJoUWqKElymFs5CNq+qSslKKSoKogyiJJarFgA7EgRyvgAAAA==",
	}
	cantDo, rsf_new, err := col.NewRsf_from_WebconsumerSelection(webdata)
	if err != nil {
		log.Fatal(err)
	}
	if cantDo {
		log.Fatal("Just cantDo now")
	}

	// ATP: runProcessingEngines() is running async, so we can wait for it to complete and then
	// show the rsf_new updated with the ProcEng results
	collection.RunnersOfProcEngs_wg.Wait()
	spew.Dump(rsf_new)
}
