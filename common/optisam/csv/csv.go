package csv

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"optisam-backend/common/optisam/logger"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// ReadDynamicCSV is for reading dynamic CSV
// Semi-colon is used as separator
// currently read the data to raw bytes array
// and unmarshal to dynamic map
// TODO Dynamic Struct
func ReadDynamicCSV(path string) ([]map[string]interface{}, error) {

	csvFile, err := os.Open(path)
	if err != nil {
		logger.Log.Error("The file is not found", zap.Error(err))
		return nil, err
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	reader.Comma = ';'
	content, _ := reader.ReadAll()

	if len(content) < 1 {
		logger.Log.Error("No Content in the file")
		return nil, err
	}

	headersArr := make([]string, 0)
	for _, headE := range content[0] {
		headersArr = append(headersArr, headE)
	}

	// Remove the header row
	content = content[1:]

	var buffer bytes.Buffer
	buffer.WriteString("[")
	for i, d := range content {
		buffer.WriteString("{")
		for j, y := range d {
			buffer.WriteString(`"` + headersArr[j] + `":`)
			_, fErr := strconv.ParseFloat(y, 32)
			_, bErr := strconv.ParseBool(y)
			if fErr == nil { // nolint: gocritic
				buffer.WriteString(y)
			} else if bErr == nil {
				buffer.WriteString(strings.ToLower(y))
			} else {
				buffer.WriteString((`"` + y + `"`))
			}
			// end of property
			if j < len(d)-1 {
				buffer.WriteString(",")
			}

		}
		// end of object of the array
		buffer.WriteString("}")
		if i < len(content)-1 {
			buffer.WriteString(",")
		}
	}

	buffer.WriteString(`]`)
	rawMessage := json.RawMessage(buffer.String())

	var mapStructure []map[string]interface{}
	err = json.Unmarshal(rawMessage, &mapStructure)
	if err != nil {
		logger.Log.Error("Failed to unmarshal", zap.Error(err))
		return nil, err
	}
	return mapStructure, nil
}
