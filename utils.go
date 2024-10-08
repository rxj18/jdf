package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
)

var INDENT = "  "

func removeANSIColors(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

func containsJSON(s string) bool {
	re := regexp.MustCompile(`\{(?:[^{}"]*|"(?:[^"\\]|\\.)*"|"(?:[^"\\]|\\.)*":(?:[^{}"]*|"(?:[^"\\]|\\.)*"|[^{}"]*)*)\}`)
	match := re.FindStringSubmatch(s)

	if len(match) == 0 {
		return false
	}

	var data interface{}
	err := json.Unmarshal([]byte(match[0]), &data)
	if err != nil {
		return false
	}
	return true
}

func getJSON(s string) string {
	var (
		braces int = 0
		start  int = -1
		end    int = -1
	)
	for index, value := range s {
		// 123 -> {
		// 125 -> }
		// 34 -> "
		if value == 123 {
			// Check for one character forward if start has not been assigned an index
			if start == -1 && s[index+1] == 34 {
				start = index
			}
			braces++
		} else if value == 125 {
			braces--
		}

		if start != -1 && braces == 0 {
			end = index
			break
		}
	}

	if end == -1 {
		end = len(s) - 1
	}
	fmt.Println(s[:start])
	return s[start : end+1]
}

func getFormattedJSON(s string) (string, error) {
	var toSend string = "{\n"
	err := getJSONMap(s, &toSend)
	if err != nil {
		return "", err
	}
	toSend = toSend[:len(toSend)-2]
	toSend = toSend + "\n}"
	return toSend, nil
}

func getJSONMap(s string, toSend *string) error {
	var data = make(map[string]interface{})
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		return err
	}

	// To maintain key order
	keys := []string{}
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := data[key]
		coloredKey := colorize("\""+key+"\"", Colors["blue"])
		delete(data, key)

		switch value.(type) {

		// ********* Case for nested objects *********
		case map[string]interface{}:
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return err
			}
			*toSend = *toSend + INDENT + coloredKey + ": {\n"

			increaseIndent()
			err = getJSONMap(string(jsonBytes), toSend)
			if err != nil {
				return err
			}

			// After the recursion stack pops
			decreaseIndent()
			*toSend = (*toSend)[:len(*toSend)-2]
			*toSend = *toSend + INDENT + "\n" + INDENT + "},\n"

		// ********* Case for array *********
		case []interface{}:
			arr, ok := value.([]interface{})
			if ok && len(arr) > 0 {
				switch arr[0].(type) {
				case map[string]interface{}:

					// build the json string
					jsonVal := ""
					for _, val := range arr {
						jsonBytes, err := json.Marshal(val)
						if err != nil {
							return err
						}

						increaseIndent()
						d, _ := getFormattedJSON(string(jsonBytes))
						decreaseIndent()
						d = d[:len(d)-1] + INDENT + d[len(d)-1:]
						jsonVal += d + ",\n" + INDENT
					} // Loop ends here

					jsonVal = jsonVal[:len(jsonVal)-2]
					*toSend = *toSend + INDENT + coloredKey + ": [" + jsonVal[:len(jsonVal)-2] + "],\n"

				case string:

					newStr := ""
					for key, val := range arr {
						if key == len(arr)-1 {
							newStr += colorize("\""+val.(string)+"\"", Colors["orange"])
							continue
						}
						newStr += colorize("\""+val.(string)+"\", ", Colors["orange"])
					}
					*toSend = *toSend + INDENT + coloredKey + ": [" + newStr + "],\n"

				default:
					newStr := ""
					for key, val := range arr {
						if key == len(arr)-1 {
							newStr += colorize(val, Colors["green"])
							continue
						}
						newStr += colorize(val, Colors["green"])
						newStr += colorize(", ", Colors["green"])
					}
					*toSend = *toSend + INDENT + coloredKey + ": [" + newStr + "],\n"
				}

			}

		// ********* Case for strings, numbers or booleans *********
		default:
			coloredValue := ""
			if _, ok := value.(string); ok {
				coloredValue = colorize("\""+value.(string)+"\"", Colors["orange"])
			} else {
				coloredValue = colorize(value, Colors["green"])
			}
			*toSend = fmt.Sprintf("%s%s%s: %v,\n", *toSend, INDENT, coloredKey, coloredValue)
		}
	}
	return nil
}

func colorize[T any](x T, color string) string {
	return fmt.Sprintf("%s%v%s", color, x, Colors["reset"])
}

func increaseIndent() {
	INDENT += "  "
}
func decreaseIndent() {
	INDENT = INDENT[:len(INDENT)-2]
}
