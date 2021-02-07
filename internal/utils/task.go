package utils

import (
	"github.com/neflyte/timetracker/internal/logger"
	"strconv"
)

func ResolveTask(arg string) (taskid uint, tasksynopsis string) {
	log := logger.GetLogger("ResolveTask")
	if arg == "" {
		return 0, ""
	}
	log.Trace().Msgf("arg=%s", arg)
	id, err := strconv.Atoi(arg)
	if err != nil {
		log.Trace().Msgf("error converting arg to number: %s; returning arg", err)
		return 0, arg
	}
	log.Trace().Msgf("returning %d", uint(id))
	return uint(id), ""
}
