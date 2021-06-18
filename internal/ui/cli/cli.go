package cli

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/rs/zerolog"
	"strings"
)

// printError prints an error message with an optional formatted message to the console
func printError(err error, format string, args ...interface{}) {
	sb := strings.Builder{}
	sb.WriteString(color.HiRedString("ERROR "))
	sb.WriteString(color.HiWhiteString(err.Error()))
	if format != "" {
		sb.WriteString(color.WhiteString(" " + fmt.Sprintf(format, args...)))
	}
	fmt.Println(sb.String())
}

// PrintAndLogError prints an error message to the console and logs it to the logger
func PrintAndLogError(log zerolog.Logger, err error, format string, args ...interface{}) {
	printError(err, format, args...)
	log.Err(err).Msgf(format, args...)
}
