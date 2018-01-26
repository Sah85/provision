package frontend

import (
	"github.com/VictorLowther/jsonpatch2"
	"github.com/digitalrebar/provision/backend"
	"github.com/digitalrebar/provision/models"
	"github.com/gin-gonic/gin"
)

// SubnetResponse returned on a successful GET, PUT, PATCH, or POST of a single subnet
// swagger:response
type SubnetResponse struct {
	// in: body
	Body *models.Subnet
}

// SubnetsResponse returned on a successful GET of all the subnets
// swagger:response
type SubnetsResponse struct {
	//in: body
	Body []*models.Subnet
}

// SubnetBodyParameter used to inject a Subnet
// swagger:parameters createSubnet putSubnet
type SubnetBodyParameter struct {
	// in: body
	// required: true
	Body *models.Subnet
}

// SubnetPatchBodyParameter used to patch a Subnet
// swagger:parameters patchSubnet
type SubnetPatchBodyParameter struct {
	// in: body
	// required: true
	Body jsonpatch2.Patch
}

// SubnetPathParameter used to name a Subnet in the path
// swagger:parameters putSubnets getSubnet putSubnet patchSubnet deleteSubnet headSubnet
type SubnetPathParameter struct {
	// in: path
	// required: true
	Name string `json:"name"`
}

// SubnetListPathParameter used to limit lists of Subnet by path options
// swagger:parameters listSubnets listStatsSubnets
type SubnetListPathParameter struct {
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
	Strategy string
	// in: query
	NextServer string
	// in: query
	Subnet string
	// in: query
	Name string
	// in: query
	Enabled string
	// in: query
	Proxy string
}

// SubnetActionsPathParameter used to find a Subnet / Actions in the path
// swagger:parameters getSubnetActions
type SubnetActionsPathParameter struct {
	// in: path
	// required: true
	Name string `json:"name"`
	// in: query
	Plugin string `json:"plugin"`
}

// SubnetActionPathParameter used to find a Subnet / Action in the path
// swagger:parameters getSubnetAction
type SubnetActionPathParameter struct {
	// in: path
	// required: true
	Name string `json:"name"`
	// in: path
	// required: true
	Cmd string `json:"cmd"`
	// in: query
	Plugin string `json:"plugin"`
}

// SubnetActionBodyParameter used to post a Subnet / Action in the path
// swagger:parameters postSubnetAction
type SubnetActionBodyParameter struct {
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

func (f *Frontend) InitSubnetApi() {
	// swagger:route GET /subnets Subnets listSubnets
	//
	// Lists Subnets filtered by some parameters.
	//
	// This will show all Subnets by default.
	//
	// You may specify:
	//    Offset = integer, 0-based inclusive starting point in filter data.
	//    Limit = integer, number of items to return
	//
	// Functional Indexs:
	//    Name = string
	//    NextServer = IP Address
	//    Subnet = CIDR Address
	//    Strategy = string
	//    Available = boolean
	//    Valid = boolean
	//    ReadOnly = boolean
	//    Enabled = boolean
	//    Proxy = boolean
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
	//    200: SubnetsResponse
	//    401: NoContentResponse
	//    403: NoContentResponse
	//    406: ErrorResponse
	f.ApiGroup.GET("/subnets",
		func(c *gin.Context) {
			f.List(c, &backend.Subnet{})
		})

	// swagger:route HEAD /subnets Subnets listStatsSubnets
	//
	// Stats of the List Subnets filtered by some parameters.
	//
	// This will return headers with the stats of the list.
	//
	// You may specify:
	//    Offset = integer, 0-based inclusive starting point in filter data.
	//    Limit = integer, number of items to return
	//
	// Functional Indexs:
	//    Name = string
	//    NextServer = IP Address
	//    Subnet = CIDR Address
	//    Strategy = string
	//    Available = boolean
	//    Valid = boolean
	//    ReadOnly = boolean
	//    Enabled = boolean
	//    Proxy = boolean
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
	f.ApiGroup.HEAD("/subnets",
		func(c *gin.Context) {
			f.ListStats(c, &backend.Subnet{})
		})

	// swagger:route POST /subnets Subnets createSubnet
	//
	// Create a Subnet
	//
	// Create a Subnet from the provided object
	//
	//     Responses:
	//       201: SubnetResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       409: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.POST("/subnets",
		func(c *gin.Context) {
			b := &backend.Subnet{}
			f.Create(c, b)
		})

	// swagger:route GET /subnets/{name} Subnets getSubnet
	//
	// Get a Subnet
	//
	// Get the Subnet specified by {name} or return NotFound.
	//
	//     Responses:
	//       200: SubnetResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	f.ApiGroup.GET("/subnets/:name",
		func(c *gin.Context) {
			f.Fetch(c, &backend.Subnet{}, c.Param(`name`))
		})

	// swagger:route HEAD /subnets/{name} Subnets headSubnet
	//
	// See if a Subnet exists
	//
	// Return 200 if the Subnet specifiec by {name} exists, or return NotFound.
	//
	//     Responses:
	//       200: NoContentResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: NoContentResponse
	f.ApiGroup.HEAD("/subnets/:name",
		func(c *gin.Context) {
			f.Exists(c, &backend.Subnet{}, c.Param(`name`))
		})

	// swagger:route PATCH /subnets/{name} Subnets patchSubnet
	//
	// Patch a Subnet
	//
	// Update a Subnet specified by {name} using a RFC6902 Patch structure
	//
	//     Responses:
	//       200: SubnetResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       406: ErrorResponse
	//       409: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.PATCH("/subnets/:name",
		func(c *gin.Context) {
			f.Patch(c, &backend.Subnet{}, c.Param(`name`))
		})

	// swagger:route PUT /subnets/{name} Subnets putSubnet
	//
	// Put a Subnet
	//
	// Update a Subnet specified by {name} using a JSON Subnet
	//
	//     Responses:
	//       200: SubnetResponse
	//       400: ErrorResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       409: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.PUT("/subnets/:name",
		func(c *gin.Context) {
			f.Update(c, &backend.Subnet{}, c.Param(`name`))
		})

	// swagger:route DELETE /subnets/{name} Subnets deleteSubnet
	//
	// Delete a Subnet
	//
	// Delete a Subnet specified by {name}
	//
	//     Responses:
	//       200: SubnetResponse
	//       401: NoContentResponse
	//       403: NoContentResponse
	//       404: ErrorResponse
	//       422: ErrorResponse
	f.ApiGroup.DELETE("/subnets/:name",
		func(c *gin.Context) {
			f.Remove(c, &backend.Subnet{}, c.Param(`name`))
		})

	subnet := &backend.Subnet{}
	pActions, pAction, pRun := f.makeActionEndpoints(subnet.Prefix(), subnet, "name")

	// swagger:route GET /subnets/{name}/actions Subnets getSubnetActions
	//
	// List subnet actions Subnet
	//
	// List Subnet actions for a Subnet specified by {name}
	//
	// Optionally, a query parameter can be used to limit the scope to a specific plugin.
	//   e.g. ?plugin=fred
	//
	//     Responses:
	//       200: ActionsResponse
	//       401: NoSubnetResponse
	//       403: NoSubnetResponse
	//       404: ErrorResponse
	f.ApiGroup.GET("/subnets/:name/actions", pActions)

	// swagger:route GET /subnets/{name}/actions/{cmd} Subnets getSubnetAction
	//
	// List specific action for a subnet Subnet
	//
	// List specific {cmd} action for a Subnet specified by {name}
	//
	// Optionally, a query parameter can be used to limit the scope to a specific plugin.
	//   e.g. ?plugin=fred
	//
	//     Responses:
	//       200: ActionResponse
	//       400: ErrorResponse
	//       401: NoSubnetResponse
	//       403: NoSubnetResponse
	//       404: ErrorResponse
	f.ApiGroup.GET("/subnets/:name/actions/:cmd", pAction)

	// swagger:route POST /subnets/{name}/actions/{cmd} Subnets postSubnetAction
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
	//       401: NoSubnetResponse
	//       403: NoSubnetResponse
	//       404: ErrorResponse
	//       409: ErrorResponse
	f.ApiGroup.POST("/subnets/:name/actions/:cmd", pRun)
}
