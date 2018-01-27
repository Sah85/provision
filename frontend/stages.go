package frontend

import (
	"github.com/VictorLowther/jsonpatch2"
	"github.com/digitalrebar/provision/backend"
	"github.com/digitalrebar/provision/models"
	"github.com/gin-gonic/gin"
)

// StageResponse returned on a successful GET, PUT, PATCH, or POST of a single stage
// swagger:response
type StageResponse struct {
	// in: body
	Body *models.Stage
}

// StagesResponse returned on a successful GET of all the stages
// swagger:response
type StagesResponse struct {
	//in: body
	Body []*models.Stage
}

// StageBodyParameter used to inject a Stage
// swagger:parameters createStage putStage
type StageBodyParameter struct {
	// in: body
	// required: true
	Body *models.Stage
}

// StagePatchBodyParameter used to patch a Stage
// swagger:parameters patchStage
type StagePatchBodyParameter struct {
	// in: body
	// required: true
	Body jsonpatch2.Patch
}

// StagePathParameter used to name a Stage in the path
// swagger:parameters putStages getStage putStage patchStage deleteStage headStage
type StagePathParameter struct {
	// in: path
	// required: true
	Name string `json:"name"`
}

// StageListPathParameter used to limit lists of Stage by path options
// swagger:parameters listStages listStatsStages
type StageListPathParameter struct {
	// in: query
	Offest int `json:"offset"`
	// in: query
	Limit int `json:"limit"`
	// in: query
	Available string
	// in: query
	Valid string
	// in: query
	ReadOnly string
	// in: query
	Name string
	// in: query
	Reboot string
	// in: query
	BootEnv string
}

// StageActionsPathParameter used to find a Stage / Actions in the path
// swagger:parameters getStageActions
type StageActionsPathParameter struct {
	// in: path
	// required: true
	Name string `json:"name"`
	// in: query
	Plugin string `json:"plugin"`
}

// StageActionPathParameter used to find a Stage / Action in the path
// swagger:parameters getStageAction
type StageActionPathParameter struct {
	// in: path
	// required: true
	Name string `json:"name"`
	// in: path
	// required: true
	Cmd string `json:"cmd"`
	// in: query
	Plugin string `json:"plugin"`
}

// StageActionBodyParameter used to post a Stage / Action in the path
// swagger:parameters postStageAction
type StageActionBodyParameter struct {
	// in: path
	// required: true
	Name string `json:"name"`
	// in: path
	// required: true
	Cmd string `json:"cmd"`
	// in: query
	Plugin string `json:"plugin"`
	// in: body
	// required: true
	Body map[string]interface{}
}

func (f *Frontend) InitStageApi() {
	// swagger:route GET /stages Stages listStages
	//
	// Lists Stages filtered by some parameters.
	//
	// This will show all Stages by default.
	//
	// You may specify:
	//    Offset = integer, 0-based inclusive starting point in filter data.
	//    Limit = integer, number of items to return
	//
	// Functional Indexs:
	//    Name = string
	//    Reboot = boolean
	//    BootEnv = string
	//    Available = boolean
	//
	// Functions:
	//    Eq(value) = Return items that are equal to value
	//    Lt(value) = Return items that are less than value
	//    Lte(value) = Return items that less than or equal to value
	//    Gt(value) = Return items that are greater than value
	//    Gte(value) = Return items that greater than or equal to value
	//    Between(lower,upper) = Return items that are inclusively between lower and upper
	//    Except(lower,upper) = Return items that are not inclusively between lower and upper
	//
	// Example:
	//    Name=fred - returns items named fred
	//    Name=Lt(fred) - returns items that alphabetically less than fred.
	//    Name=Lt(fred)&Available=true - returns items with Name less than fred and Available is true
	//
	// Responses:
	//    200: StagesResponse
	//    401: NoContentResponse
	//    403: NoContentResponse
	//    406: ErrorResponse
	f.ApiGroup.GET("/stages",
		func(c *gin.Context) {
			f.List(c, &backend.Stage{})
		})

	// swagger:route HEAD /stages Stages listStatsStages
	//
	// Stats of the List Stages filtered by some parameters.
	//
	// This will return headers with the stats of the list.
	//
	// You may specify:
	//    Offset = integer, 0-based inclusive starting point in filter data.
	//    Limit = integer, number of items to return
	//
	// Functional Indexs:
	//    Name = string
	//    Reboot = boolean
	//    BootEnv = string
	//    Available = boolean
	//
	// Functions:
	//    Eq(value) = Return items that are equal to value
	//    Lt(value) = Return items that are less than value
	//    Lte(value) = Return items that less than or equal to value
	//    Gt(value) = Return items that are greater than value
	//    Gte(value) = Return items that greater than or equal to value
	//    Between(lower,upper) = Return items that are inclusively between lower and upper
	//    Except(lower,upper) = Return items that are not inclusively between lower and upper
	//
	// Example:
	//    Name=fred - returns items named fred
	//    Name=Lt(fred) - returns items that alphabetically less than fred.
	//    Name=Lt(fred)&Available=true - returns items with Name less than fred and Available is true
	//
	// Responses:
	//    200: NoContentResponse
	//    401: NoContentResponse
	//    403: NoContentResponse
	//    406: ErrorResponse
	f.ApiGroup.HEAD("/stages",
		func(c *gin.Context) {
			f.ListStats(c, &backend.Stage{})
		})

	// swagger:route POST /stages Stages createStage
	//
	// Create a Stage
	//
	// Create a Stage from the provided object
	//
	//     Responses:
	//       201: StageResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       409: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.POST("/stages",
		func(c *gin.Context) {
			b := &backend.Stage{}
			f.Create(c, b)
		})
	// swagger:route GET /stages/{name} Stages getStage
	//
	// Get a Stage
	//
	// Get the Stage specified by {name} or return NotFound.
	//
	//     Responses:
	//       200: StageResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	f.ApiGroup.GET("/stages/:name",
		func(c *gin.Context) {
			f.Fetch(c, &backend.Stage{}, c.Param(`name`))
		})

	// swagger:route HEAD /stages/{name} Stages headStage
	//
	// See if a Stage exists
	//
	// Return 200 if the Stage specifiec by {name} exists, or return NotFound.
	//
	//     Responses:
	//       200: NoContentResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: NoContentResponse
	f.ApiGroup.HEAD("/stages/:name",
		func(c *gin.Context) {
			f.Exists(c, &backend.Stage{}, c.Param(`name`))
		})

	// swagger:route PATCH /stages/{name} Stages patchStage
	//
	// Patch a Stage
	//
	// Update a Stage specified by {name} using a RFC6902 Patch structure
	//
	//     Responses:
	//       200: StageResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       406: ErrorResponse
	//       409: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.PATCH("/stages/:name",
		func(c *gin.Context) {
			f.Patch(c, &backend.Stage{}, c.Param(`name`))
		})

	// swagger:route PUT /stages/{name} Stages putStage
	//
	// Put a Stage
	//
	// Update a Stage specified by {name} using a JSON Stage
	//
	//     Responses:
	//       200: StageResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       409: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.PUT("/stages/:name",
		func(c *gin.Context) {
			f.Update(c, &backend.Stage{}, c.Param(`name`))
		})

	// swagger:route DELETE /stages/{name} Stages deleteStage
	//
	// Delete a Stage
	//
	// Delete a Stage specified by {name}
	//
	//     Responses:
	//       200: StageResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       409: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.DELETE("/stages/:name",
		func(c *gin.Context) {
			f.Remove(c, &backend.Stage{}, c.Param(`name`))
		})

	stage := &backend.Stage{}
	pActions, pAction, pRun := f.makeActionEndpoints(stage.Prefix(), stage, "name")

	// swagger:route GET /stages/{name}/actions Stages getStageActions
	//
	// List stage actions Stage
	//
	// List Stage actions for a Stage specified by {name}
	//
	// Optionally, a query parameter can be used to limit the scope to a specific plugin.
	//   e.g. ?plugin=fred
	//
	//     Responses:
	//       200: ActionsResponse
	//       401: NoStageResponse
	//       403: NoStageResponse
	//       404: ErrorResponse
	f.ApiGroup.GET("/stages/:name/actions", pActions)

	// swagger:route GET /stages/{name}/actions/{cmd} Stages getStageAction
	//
	// List specific action for a stage Stage
	//
	// List specific {cmd} action for a Stage specified by {name}
	//
	// Optionally, a query parameter can be used to limit the scope to a specific plugin.
	//   e.g. ?plugin=fred
	//
	//     Responses:
	//       200: ActionResponse
	//       400: ErrorResponse
	//       401: NoStageResponse
	//       403: NoStageResponse
	//       404: ErrorResponse
	f.ApiGroup.GET("/stages/:name/actions/:cmd", pAction)

	// swagger:route POST /stages/{name}/actions/{cmd} Stages postStageAction
	//
	// Call an action on the node.
	//
	// Optionally, a query parameter can be used to limit the scope to a specific plugin.
	//   e.g. ?plugin=fred
	//
	//
	//     Responses:
	//       400: ErrorResponse
	//       200: ActionPostResponse
	//       401: NoStageResponse
	//       403: NoStageResponse
	//       404: ErrorResponse
	//       409: ErrorResponse
	f.ApiGroup.POST("/stages/:name/actions/:cmd", pRun)
}
