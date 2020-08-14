// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package fileworker

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	gendb "optisam-backend/dps-service/pkg/repository/v1/postgres/db"
	"optisam-backend/dps-service/pkg/worker/constants"
	"optisam-backend/dps-service/pkg/worker/models"
	//"github.com/pkg/profile"
)

type worker struct {
	id string
	*workerqueue.Queue
	*gendb.Queries
}

//NewWorker give worker object
func NewWorker(id string, queue *workerqueue.Queue, db *sql.DB) *worker {
	return &worker{id: id, Queue: queue, Queries: gendb.New(db)}
}

//ID gives unique id of worker
func (w *worker) ID() string {
	return w.id
}

//DoWork tell the functionality of worker
func (w *worker) DoWork(ctx context.Context, j *job.Job) error {
	//defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	dataFromJob := gendb.UploadedDataFile{}
	var data models.FileData

	err := json.Unmarshal(j.Data, &dataFromJob)
	if err != nil {
		log.Println("Failed to unmarshal the file type job data , err :", err)
		return err
	}
	defer archiveFile(dataFromJob.FileName, dataFromJob.UploadID)

	dataToUpdate := gendb.UpdateFileStatusParams{
		UploadID: dataFromJob.UploadID,
		FileName: dataFromJob.FileName,
		Status:   gendb.UploadStatusINPROGRESS,
	}

	err = w.Queries.UpdateFileStatus(ctx, dataToUpdate)
	if err != nil {
		log.Println("Failed to update status , err ", err)
		return err
	}

	data, err = fileProcessing(dataFromJob)
	if err != nil {
		log.Println("Failed to process the file ", dataFromJob.FileName, " err : ", err)
		dataToUpdate.Status = gendb.UploadStatusFAILED
		dbErr := w.Queries.UpdateFileStatus(ctx, dataToUpdate)
		if dbErr != nil {
			log.Println("Failed to update the status of file ", dataFromJob.FileName, " , err :", err)
			return dbErr
		}
		return err
	}
	log.Println(" <<<<>>>>>>>>>>>> File processed ", dataFromJob.FileName)

	err = w.Queries.UpdateFileTotalRecord(ctx, gendb.UpdateFileTotalRecordParams{
		FileName:     dataFromJob.FileName,
		UploadID:     dataFromJob.UploadID,
		TotalRecords: data.TotalCount})
	if err != nil {
		log.Println("Failed to update total Records in DB for file ", dataFromJob.FileName, " err :", err)
		return err
	}

	//log.Printf(" %s  file's complete data  from file: %+v", dataFromJob.FileName, data)
	jobs, err := createAPITypeJobs(data)
	lenJ := len(jobs)
	log.Println(" <<<<>>>>>>>>>>>> Jobs created in memory ", dataFromJob.FileName, lenJ)
	for _, job := range jobs {
		_, err = w.Queue.PushJob(ctx, job, constants.APIWORKER)
		if err != nil {
			log.Println("Failed to push api type jobs  , err :", err)
			return err
		}
	}
	log.Println(" <<<<>>>>>>>>>>>> Jobs Pushed ", dataFromJob.FileName)
	dataToUpdate.Status = gendb.UploadStatusCOMPLETED
	err = w.Queries.UpdateFileStatus(ctx, dataToUpdate)
	if err != nil {
		log.Println("Failed to update status , err ", err)
		return err
	}
	return nil
}
