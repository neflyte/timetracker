package models

// TaskList represents a slice of tasks
type TaskList []Task

// ToSliceIntf converts the slice of tasks into a slice of interface{}
func (tl TaskList) ToSliceIntf() []interface{} {
	taskListIntf := make([]interface{}, len(tl))
	for idx := range tl {
		taskIntf, ok := tl[idx].(interface{})
		if ok {
			taskListIntf[idx] = taskIntf
		}
	}
	return taskListIntf
}

// Index returns the index of the specified Task in the list or -1 if the Task is not found
func (tl TaskList) Index(task Task) int {
	if task == nil {
		return -1
	}
	for idx := range tl {
		if tl[idx].Equals(task) {
			return idx
		}
	}
	return -1
}

// Contains tests whether the supplied Task is present in the list
func (tl TaskList) Contains(task Task) bool {
	if task == nil {
		return false
	}
	return tl.Index(task) > -1
}

// Names returns a slice of strings containing the DisplayString for each Task
func (tl TaskList) Names() []string {
	names := make([]string, len(tl))
	for idx := range tl {
		names[idx] = tl[idx].DisplayString()
	}
	return names
}

// TaskListFromSliceIntf converts a slice of interface{} into a slice of tasks (TaskList).
// Elements that are not TaskList are excluded from the result.
func TaskListFromSliceIntf(taskListIntf []interface{}) TaskList {
	if taskListIntf == nil {
		return make(TaskList, 0)
	}
	tasks := make(TaskList, len(taskListIntf))
	for idx := range taskListIntf {
		task, ok := taskListIntf[idx].(Task)
		if ok {
			tasks[idx] = task
		}
	}
	return tasks
}
