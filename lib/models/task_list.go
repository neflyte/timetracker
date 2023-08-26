package models

// TaskList represents a list of tasks
type TaskList []Task

// TaskListFromSliceIntf converts a slice of interface{} into a slice of tasks (TaskList).
// Elements that are not TaskList are excluded from the result.
func TaskListFromSliceIntf(taskListIntf []interface{}) TaskList {
	tasks := make(TaskList, 0)
	for idx := range taskListIntf {
		task, ok := taskListIntf[idx].(Task)
		if ok {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// TaskListToSliceIntf converts a slice of tasks (TaskList) into a slice of interface{}
func TaskListToSliceIntf(taskList TaskList) []interface{} {
	taskListIntf := make([]interface{}, 0)
	for idx := range taskList {
		taskIntf, ok := taskList[idx].(interface{})
		if ok {
			taskListIntf = append(taskListIntf, taskIntf)
		}
	}
	return taskListIntf
}
