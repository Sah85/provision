package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/digitalrebar/provision/models"
	"github.com/spf13/cobra"
)

func init() {
	addRegistrar(registerMachine)
}

var actuallyPowerThings = true

func registerMachine(app *cobra.Command) {
	op := &ops{
		name:       "machines",
		singleName: "machine",
		example:    func() models.Model { return &models.Machine{} },
	}
	op.addCommand(&cobra.Command{
		Use:   "stage [id] [stage]",
		Short: fmt.Sprintf("Set the machine's stage"),
		Long:  `Helper function to update the machine's stage.`,
		Args: func(c *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("%v requires 2 arguments", c.UseLine())
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			m, err := op.refOrFill(args[0])
			if err != nil {
				return generateError(err, "Failed to fetch %v: %v", op.singleName, args[0])
			}
			clone := models.Clone(m).(*models.Machine)
			clone.Stage = args[1]
			req := session.Req().ParanoidPatch().PatchTo(m, clone)
			if force {
				req.Params("force", "true")
			}
			if err := req.Do(&clone); err != nil {
				return err
			}
			return prettyPrint(clone)
		},
	})
	tasks := &cobra.Command{
		Use:   "tasks",
		Short: "Access task manipulation for machines",
	}
	tasks.AddCommand(&cobra.Command{
		Use:   "add [id] [at [offset]] [task...]",
		Short: "Add tasks to the task list for [id]",
		Long: `You may omit "at [offset]" to indicate that the task should be appended to the
end of the task list.  Otherwise, [offset] 0 indicates that the tasks
should be inserted immediately after the current task. Negative numbers
are not accepted.`,
		Args: func(c *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("%v requires at least an id and one task", c.UseLine())
			}
			if args[1] == "at" {
				if len(args) < 4 {
					return fmt.Errorf("%v requires at least 3 arguments when specifying an offset", c.UseLine())
				}
				if _, err := strconv.Atoi(args[2]); err != nil {
					return fmt.Errorf("%v: %s is not a number", args[2])
				}
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			var offset = -1
			var tasks []string
			if args[1] == "at" {
				offset, _ = strconv.Atoi(args[2])
				tasks = args[3:]
			} else {
				tasks = args[1:]
			}
			obj, err := op.refOrFill(id)
			if err != nil {
				return err
			}
			m := models.Clone(obj).(*models.Machine)
			if err := m.AddTasks(offset, tasks...); err != nil {
				generateError(err, "Cannot add tasks")
			}
			if err := session.Req().PatchTo(obj, m).Do(&m); err != nil {
				return err
			}
			return prettyPrint(m)
		},
	})
	tasks.AddCommand(&cobra.Command{
		Use:   "del [task...]",
		Short: "Remove tasks from the mutable part of the task list",
		Long: `Each entry in [task...] will remove at most one instance of it from the
machine task list.  If you want to remove more than one, you need to
pass in more than one task.`,
		Args: func(c *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("%v requires at least an id and one task", c.UseLine())
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			tasks := args[1:]
			obj, err := op.refOrFill(id)
			if err != nil {
				return err
			}
			m := models.Clone(obj).(*models.Machine)
			m.DelTasks(tasks...)
			if err := session.Req().PatchTo(obj, m).Do(&m); err != nil {
				return err
			}
			return prettyPrint(m)
		},
	})
	op.addCommand(tasks)
	var exitOnFailure = false
	processJobs := &cobra.Command{
		Use:   "processjobs [id]",
		Short: "For the given machine, process pending jobs until done.",
		Long: `
For the provided machine, identified by UUID, process the task list on
that machine until an error occurs or all jobs are complete.  Upon
completion, optionally wait for additional jobs as specified by
the stage runner wait flag.
`,
		Args: func(c *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("%v requires 1 argument", c.UseLine())
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			uuid := args[0]
			m := &models.Machine{}
			if err := session.FillModel(m, uuid); err != nil {
				return err
			}

			return session.Agent(m, false, exitOnFailure, actuallyPowerThings, os.Stdout)
		},
	}
	processJobs.Flags().BoolVar(&exitOnFailure, "exit-on-failure", false, "Exit on failure of a task")
	op.addCommand(processJobs)
	op.command(app)
}
