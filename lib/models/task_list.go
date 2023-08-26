package models

// TaskList represents a list of tasks
type TaskList []Task

// TaskListFromSliceIntf converts a slice of interface{} into a slice of tasks (TaskList)
func TaskListFromSliceIntf(taskListIntf []interface{}) TaskList {
	tasks := make(TaskList, 0)
	if taskListIntf == nil {
		return tasks
	}
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
	if taskList == nil {
		return taskListIntf
	}
	for idx := range taskList {
		taskListIntf = append(taskListIntf, taskList[idx].(interface{}))
	}
	return taskListIntf
}
