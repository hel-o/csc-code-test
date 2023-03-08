package apis

import (
	"csc-code-test/internal/settings"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func init() {
	settings.LoadSettings()
	InitializeJobQueueManager()
}

func TestPushUnauthorizedJob(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()

	newJobsPayload := `[{"name": "1st job", "data": [1, 2, 3]}]`

	reqPostJob := httptest.NewRequest(http.MethodPost, "/api/v1/jobs", strings.NewReader(newJobsPayload))
	reqPostJob.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	ctx := e.NewContext(reqPostJob, rec)

	if assert.NoError(t, JobsPOST(ctx)) {
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Equal(t, `{"error":"Unauthorized to access this resource","status":401}`, strings.TrimSpace(rec.Body.String()))
	}
}

/*
We are using a global shared variable for handling the job queues.
It's hardly to test the different API endpoints in separate functions due to the global shared variable behavior so,
I'm using one single test function:
*/
func TestPushJobsAndProcess(t *testing.T) {
	JobQueueManagerShared.StartWorker()

	e := echo.New()

	newJobsPayload := `[{"name": "1st job", "data": [1, 2, 3]},{"name": "2nd job", "data": [10, 20, 30]},{"name": "3rd job", "data": [100, 200, 300]}]`

	reqPostJob := httptest.NewRequest(http.MethodPost, "/api/v1/jobs", strings.NewReader(newJobsPayload))
	reqPostJob.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	reqPostJob.Header.Set(echo.HeaderAuthorization, "allow")
	recPostJob := httptest.NewRecorder()

	ctx := e.NewContext(reqPostJob, recPostJob)

	if assert.NoError(t, JobsPOST(ctx)) {
		assert.Equal(t, http.StatusCreated, recPostJob.Code)
	}

	// Test - list all pending queued jobs:

	reqListAllJobs := httptest.NewRequest(http.MethodGet, "/api/v1/jobs", nil)
	recListAllJobs := httptest.NewRecorder()
	ctxListAllJobs := e.NewContext(reqListAllJobs, recListAllJobs)

	if assert.NoError(t, JobsGET(ctxListAllJobs)) {
		assert.Equal(t, http.StatusOK, recListAllJobs.Code)

		allJobsPayload := `[{"id":1,"name":"1st job","data":[1,2,3],"status":"pending","result":0},{"id":2,"name":"2nd job","data":[10,20,30],"status":"pending","result":0},{"id":3,"name":"3rd job","data":[100,200,300],"status":"pending","result":0}]`

		assert.Equal(t, allJobsPayload, strings.TrimSpace(recListAllJobs.Body.String()))
	}

	// wait the default time for processing one job:

	time.Sleep(time.Duration(settings.Settings.SecondsToWait+1) * time.Second)

	// Test - list all remaining pending jobs:

	reqListPendingJob := httptest.NewRequest(http.MethodGet, "/api/v1/jobs", nil)
	recListPendingJob := httptest.NewRecorder()
	ctxListPendingJob := e.NewContext(reqListPendingJob, recListPendingJob)
	ctxListPendingJob.SetPath("/:status")
	ctxListPendingJob.SetParamNames("status")
	ctxListPendingJob.SetParamValues("pending")

	if assert.NoError(t, JobsGET(ctxListPendingJob)) {
		assert.Equal(t, http.StatusOK, recListPendingJob.Code)

		pendingJobsPayload := `[{"id":2,"name":"2nd job","data":[10,20,30],"status":"pending","result":0},{"id":3,"name":"3rd job","data":[100,200,300],"status":"pending","result":0}]`

		assert.Equal(t, pendingJobsPayload, strings.TrimSpace(recListPendingJob.Body.String()))
	}

	// Test - list all processed jobs:

	reqListProcessedJob := httptest.NewRequest(http.MethodGet, "/api/v1/jobs", nil)
	recListProcessedJob := httptest.NewRecorder()
	ctxListProcessedJob := e.NewContext(reqListProcessedJob, recListProcessedJob)
	ctxListProcessedJob.SetPath("/:status")
	ctxListProcessedJob.SetParamNames("status")
	ctxListProcessedJob.SetParamValues("processed")

	if assert.NoError(t, JobsGET(ctxListProcessedJob)) {
		assert.Equal(t, http.StatusOK, recListProcessedJob.Code)

		processedJobsPayload := `[{"id":1,"name":"1st job","data":[1,2,3],"status":"processed","result":6}]`

		assert.Equal(t, processedJobsPayload, strings.TrimSpace(recListProcessedJob.Body.String()))
	}

	// Test - list both jobs with status pending and processed:

	reqListMixAllJobs := httptest.NewRequest(http.MethodGet, "/api/v1/jobs", nil)
	recListMixAllJobs := httptest.NewRecorder()
	ctxListMixAllJobs := e.NewContext(reqListMixAllJobs, recListMixAllJobs)

	if assert.NoError(t, JobsGET(ctxListMixAllJobs)) {
		assert.Equal(t, http.StatusOK, recListMixAllJobs.Code)

		listMixJobsPayload := `[{"id":1,"name":"1st job","data":[1,2,3],"status":"processed","result":6},{"id":2,"name":"2nd job","data":[10,20,30],"status":"pending","result":0},{"id":3,"name":"3rd job","data":[100,200,300],"status":"pending","result":0}]`

		assert.Equal(t, listMixJobsPayload, strings.TrimSpace(recListMixAllJobs.Body.String()))
	}

	JobQueueManagerShared.StopWorker()
}
