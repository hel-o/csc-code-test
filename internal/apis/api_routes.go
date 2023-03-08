package apis

import (
	"csc-code-test/internal/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
)

func RouteApiJobs(group *echo.Group) {
	group.GET("/jobs", JobsGET)
	group.GET("/jobs/:status", JobsGET)
	group.POST("/jobs", JobsPOST)
}

func JobsGET(ctx echo.Context) error {
	jobStatus := ctx.Param("status")

	itemsJob := JobQueueManagerShared.ListJobs(jobStatus)

	return ctx.JSON(http.StatusOK, itemsJob)
}

type NewJobFormModel struct {
	Name string `json:"name"`
	Data []int  `json:"data"`
}

var InvalidRequest struct {
	ErrorMsg string `json:"error"`
	Status   int    `json:"status"`
}

func JobsPOST(ctx echo.Context) error {
	authHeader := ctx.Request().Header.Get(echo.HeaderAuthorization)

	if authHeader != "allow" {

		InvalidRequest.ErrorMsg = "Unauthorized to access this resource"
		InvalidRequest.Status = http.StatusUnauthorized

		return ctx.JSON(http.StatusUnauthorized, InvalidRequest)
	}

	var itemsJobs []NewJobFormModel

	if err := ctx.Bind(&itemsJobs); err != nil {
		logger.Logger.Error("err bind", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, nil)
	}

	for _, item := range itemsJobs {
		JobQueueManagerShared.PushJobTask(item.Name, item.Data)
	}

	return ctx.JSON(http.StatusCreated, map[string]string{})
}
