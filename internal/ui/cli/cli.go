package cli

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/jszwec/csvutil"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/rs/zerolog"
)

var (
	packageLogger = logger.GetPackageLogger("cli") // nolint:unused
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

func PrintCSV(log zerolog.Logger, obj interface{}) {
	csvOut, err := csvutil.Marshal(obj)
	if err != nil {
		log.Err(err).Msg("unable to marshal object to CSV")
		return
	}
	fmt.Println(string(csvOut))
}

func PrintJSON(log zerolog.Logger, obj interface{}) {
	jsonOut, err := json.Marshal(obj)
	if err != nil {
		log.Err(err).Msg("unable to marshal object to JSON")
		return
	}
	fmt.Println(string(jsonOut))
}

func PrintXML(log zerolog.Logger, obj interface{}) {
	xmlOut, err := xml.Marshal(obj)
	if err != nil {
		log.Err(err).Msg("unable to marshal object to XML")
		return
	}
	fmt.Println(string(xmlOut))
}
