package printutils

import (
	"io"

	"encoding/json"
)

func PrettyPrint(data interface{}) {
	content, _ := json.MarshalIndent(data, "", "    ")
	println(string(content))
}

func PrettyPrintX(data interface{}, writer io.Writer) {
	content, _ := json.MarshalIndent(data, "", "    ")
	writer.Write(content)
}
