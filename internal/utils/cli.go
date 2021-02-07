package utils

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/rs/zerolog"
)

func PrintError(description string, err error) {
	fmt.Println(color.HiRedString("%s: ", description), color.WhiteString(err.Error()))
}

func PrintAndLogError(description string, err error, log zerolog.Logger) {
	PrintError(description, err)
	log.Err(err).Msg(description)
}
