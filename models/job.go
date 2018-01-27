package models

import (
	"fmt"
	"time"

	"github.com/pborman/uuid"
)

// Job Action is something that job runner will need to do.
// If path is specified, then the runner will place the contents into that location.
// If path is not specified, then the runner will attempt to bash exec the contents.
// swagger:model
type JobAction struct {
	// required: true
	Name string
	// required: true
	Path string
	// required: true
	Content string
}

// swagger:model
type Job struct {
	Validation
	Access
	Meta
	// The UUID of the job.  The primary key.
	// required: true
	// swagger:strfmt uuid
	Uuid uuid.UUID
	// The UUID of the previous job to run on this machine.
	// swagger:strfmt uuid
	Previous uuid.UUID
	// The machine the job was created for.  This field must be the UUID of the machine.
	// required: true
	// swagger:strfmt uuid
	Machine uuid.UUID
	// The task the job was created for.  This will be the name of the task.
	// read only: true
	Task string
	// The stage that the task was created in.
	// read only: true
	Stage string
	// The state the job is in.  Must be one of "created", "running", "failed", "finished", "incomplete"
	// required: true
	State string
	// The final disposition of the job.
	// Can be one of "reboot","poweroff","stop", or "complete"
	// Other substates may be added as time goes on
	ExitState string
	// The time the job entered running.
	StartTime time.Time
	// The time the job entered failed or finished.
	EndTime time.Time
	// required: true
	Archived bool
	// Whether the job is the "current one" for the machine or if it has been superceded.
	//
	// required: true
	Current bool
}

func (j *Job) Validate() {
	j.AddError(ValidName("Invalid Task", j.Task))
	j.AddError(ValidName("Invalid Stage", j.Stage))
	switch j.State {
	case "created", "running", "incomplete":
	case "failed", "finished":
	default:
		j.AddError(fmt.Errorf("Invalid State `%s`", j.State))
	}
	if j.ExitState != "" {
		switch j.ExitState {
		case "reboot", "poweroff", "stop", "complete", "failed":
		default:
			j.AddError(fmt.Errorf("Invalid ExitState `%s`", j.ExitState))
		}
	}
}

func (j *Job) Prefix() string {
	return "jobs"
}

func (j *Job) Key() string {
	return j.Uuid.String()
}

func (j *Job) Fill() {
	if j.Meta == nil {
		j.Meta = Meta{}
	}
	j.Validation.fill()
}

func (j *Job) AuthKey() string {
	return j.Machine.String()
}

func (b *Job) SliceOf() interface{} {
	s := []*Job{}
	return &s
}

func (b *Job) ToModels(obj interface{}) []Model {
	items := obj.(*[]*Job)
	res := make([]Model, len(*items))
	for i, item := range *items {
		res[i] = Model(item)
	}
	return res
}

func (b *Job) CanHaveActions() bool {
	return true
}
