package main

import (
	"fmt"
	// "unicode/utf8"
	"os"
	// "os/exec"
	"strings"
	"encoding/json"
	// "encoding/hex"
	// "io"
	// "net/http"
    // "net/url"
	// "errors"
	// "bufio"
	"strconv"
	// "sync"
	// "path/filepath"
)

const (
	tag string = "[Mard]"	
	finalM3u8Name string = "target.m3u8"
	m3u8Name string = "targetMaster.m3u8"
)

var (
	err error
	errw error

	rType string
	bType string

	qchoice int

	gqdata map[string]any

	m3u8PathUrl string
	m3u8Url string
	outputFile string
)

func main() {
	fmt.Println(tag, "Made by Totem - M.A.R.D 1.0.0 - 2026")

	if len(os.Args) < 3 {
		fmt.Println(tag, "Usage: mard <type> <service> <m3u8_url> <output_file>")
		return
	}

	// Load config
	fmt.Println(tag, "Loading config...")
	err = loadConfig()
	if err != nil {
		fmt.Println(tag, "Config Error: ", err)
		return
	} else if strings.TrimSpace(asa) == "" || asa == "YOUR_APPSCRIPT_API_KEY" {
		fmt.Println(tag, "Config Error: AppScript api code is missing")
		return
	}

	// Check download directory
	err = downloadDirectoryCheck()
	if err != nil {
		fmt.Println(tag, "Create download folder error:", err)
		return
	}

	// CMD
	bType = os.Args[1]
	rType = os.Args[2]
	arg3 := os.Args[3]
	if bType != "stream" && bType != "download" && bType != "watch" {
		fmt.Println(tag, "Invalid type. download | stream | watch")
	}
	if strings.EqualFold(rType, "twitch") {
		errw, err, m3u8Url = TgetStreamM3U8(arg3)
		
		if err != nil {
			return
		}
	} else {
		m3u8Url = arg3
	}

	if bType == "download" {
		outputFile = os.Args[4]
	}

	// Delete files
	deleteTs()

	// Download the master M3U8
	fmt.Println(tag, "Downloading master M3U8...")
	errw, err = getM3U8(m3u8Url, m3u8Name)
	if err != nil {
		switch errw.Error() {
			case "gm:1":
				fmt.Println(tag, "Connection error:", err)
				return
			case "gm:2":
				fmt.Println(tag, "bodyBytes error:", err)
				return
			case "gm:3":
				fmt.Println(tag, "AppScript error:", err)
				return
			case "gm:4":
				fmt.Println(tag, "Request error:", err)
				return
			case "gm:5":
				fmt.Println(tag, "Error: Its not M3U8 file!")
				return
			case "gm:6":
				fmt.Println(tag, "Write file error:", err)
				return
			default:
				fmt.Println(tag, err)
				return
		}
	}
	fmt.Println(tag, "Master M3U8 downloaded")

	// Set the base M3U8 URL
	m3u8PathUrl = strings.Split(m3u8Url, strings.Split(m3u8Url, "/")[len(strings.Split(m3u8Url, "/")) - 1])[0]

	// Detect available qualities
	fmt.Println(tag, "Checking available qualities...")
	errw, err, qualitiesby := getQualities()
	if err != nil { 
		switch errw.Error() {
			case "gq:1":
				fmt.Println(tag, "Error: M3U8 not found! (gq:1)")
				return
			case "gq:2":
				fmt.Println(tag, "Open M3U8 file error:", err)
				return
			case "gq:3":
				fmt.Println(tag, "Read file error:", err)
				return
			case "gq:4":
				fmt.Println(tag, "Quality json marshal error", err)
				return
			case "gq:5":
				fmt.Println(tag, "Its not a Master M3U8 file!")
				finalM3u8Url = m3u8Url
			default:
				fmt.Println(tag, err)
				return
		}
	} else {
		fmt.Println(tag, "Select one of the qualities below by entering the number next to it")

		err := json.Unmarshal(qualitiesby, &gqdata)
		if err != nil {
			return
		}

		qualities := gqdata["qualities"].([]any)

		for i, q := range qualities {
			item := q.(map[string]any)

			fmt.Printf("[%d] %s - %s\n",
				i+1,
				item["resolution"],
				item["bandwidth"],
			)
		}

		fmt.Print(tag, " Enter quality number: ")
		_, err = fmt.Scanln(&qchoice)
		if err != nil {
			return
		}

		if qchoice < 1 || qchoice > len(qualities) {
			fmt.Println(tag, "Invalid choice")
			return
		}

		selectedQuality := qualities[qchoice-1].(map[string]any)

		fmt.Println(tag, "URL:", selectedQuality["url"])
		fmt.Println(tag, "Resolution:", selectedQuality["resolution"])
		fmt.Println(tag, "Bandwidth:", selectedQuality["bandwidth"])

		if strings.Contains(selectedQuality["url"].(string), "http://") || strings.Contains(selectedQuality["url"].(string), "https://") {
			finalM3u8Url = selectedQuality["url"].(string)
			m3u8PathUrl = strings.Split(finalM3u8Url, strings.Split(finalM3u8Url, "/")[len(strings.Split(finalM3u8Url, "/")) - 1])[0]
		} else {
			finalM3u8Url = m3u8PathUrl + selectedQuality["url"].(string)

			if strings.Contains(selectedQuality["url"].(string), "/") {
				m3u8PathUrlup := strings.Split(selectedQuality["url"].(string), "/")[len(strings.Split(selectedQuality["url"].(string), "/")) - 1]
				m3u8PathUrlup = strings.Split(selectedQuality["url"].(string), m3u8PathUrlup)[0]
				m3u8PathUrl = m3u8PathUrl + m3u8PathUrlup
			}
		}
		
	}

	if bType == "download" {
		// Download the selected M3U8
		fmt.Println(tag, "Downloading target M3U8...")
		errw, err = getM3U8(finalM3u8Url, finalM3u8Name)
		if err != nil { 
			switch errw.Error() {
				case "gm:1":
					fmt.Println(tag, "Connection error:", err)
					return
				case "gm:2":
					fmt.Println(tag, "bodyBytes error:", err)
					return
				case "gm:3":
					fmt.Println(tag, "AppScript error:", err)
					return
				case "gm:4":
					fmt.Println(tag, "Request error:", err)
					return
				case "gm:5":
					fmt.Println(tag, "Error: Its not M3U8 file!")
					return
				case "gm:6":
					fmt.Println(tag, "Write file error:", err)
					return
				default:
					fmt.Println(tag, err)
					return
			}
		}
		fmt.Println(tag, "Target M3U8 downloaded")

		// Parse TS segments
		fmt.Println(tag, "Parsing target M3U8...")
		errw, err, tsList, tsNum, finalText := getTsNum(finalM3u8Name)
		finalM3u8TsNum = tsNum
		if err != nil {
			switch errw.Error() {
				case "gtn:1":
					fmt.Println(tag, "Error: M3U8 not found! (gtn:1)")
					return
				case "gtn:2":
					fmt.Println(tag, "M3U8 Read Error:", err)
					return
				case "gtn:3":
					fmt.Println(tag, "Batch Create Error:", err)
					return
				default:
					fmt.Println(tag, err)
					return
			}
		}
		if tsNum <= 0 {
			fmt.Println(tag, "Error: Target M3U8 is empty!")
			return
		}
		fmt.Println(tag, strconv.Itoa(tsNum) + " ts found")

		//Write the final M3U8
		err = finalFileWriter(finalText)
		if err != nil {
			fmt.Println(tag, "Write final file error", err)
			return
		}
		
		//Get ts'
		err = gat(tsList)
		if err != nil {
			fmt.Println(tag, "Batch unmarshal error: ", err)
		}

		//Convert the ts'
		fmt.Println(tag, "M3U8 is converting...")
		err = ffmpeg(dp + "final" + finalM3u8Name, outputFile)
		if err != nil {
			fmt.Println(tag, "Convert error:", err)
			fmt.Println(tag, "Use the ffmpeg or other apps for convert created M3U8 file to MP4 - Filename: " + "final" + finalM3u8Name)
			return
		} else {
			fmt.Println(tag, "M3U8 converted")
		}

		//Delete files
		deleteTs()
		fmt.Println(tag, "Ts' deleted")
	} else if bType == "stream" {
		getOk := true
		slisten := false
		for getOk == true {
			fmt.Println(tag, "Updating target M3U8...")
			errw, err = getM3U8(finalM3u8Url, finalM3u8Name)
			if err != nil { 
				switch errw.Error() {
					case "gm:1":
						fmt.Println(tag, "Connection error:", err)
						return
					case "gm:2":
						fmt.Println(tag, "bodyBytes error:", err)
						return
					case "gm:3":
						fmt.Println(tag, "AppScript error:", err)
						return
					case "gm:4":
						fmt.Println(tag, "Request error:", err)
						return
					case "gm:5":
						fmt.Println(tag, "Error: Its not M3U8 file!")
						return
					case "gm:6":
						fmt.Println(tag, "Write file error:", err)
						return
					default:
						fmt.Println(tag, err)
						return
				}
			}
			fmt.Println(tag, "Target M3U8 updated")

			// Parse TS segments
			fmt.Println(tag, "Parsing target M3U8...")
			errw, err, tsList, tsNum, finalText := getTsNum(finalM3u8Name)
			finalM3u8TsNum = tsNum
			if err != nil {
				switch errw.Error() {
					case "gtn:1":
						fmt.Println(tag, "Error: M3U8 not found! (gtn:1)")
						return
					case "gtn:2":
						fmt.Println(tag, "M3U8 Read Error:", err)
						return
					case "gtn:3":
						fmt.Println(tag, "Batch Create Error:", err)
						return
					case "gtn:4":
						fmt.Println(tag, "Write final file error", err)
						return
					default:
						fmt.Println(tag, err)
						return
				}
			}
			if tsNum <= 0 {
				fmt.Println(tag, "Error: Target M3U8 is empty!")
				return
			}

			if lowA {
				//Write the final M3U8
				err = finalFileWriter(finalText)
				if err != nil {
					fmt.Println(tag, "Write final file error", err)
					return
				}
			}

			//Port listen
			if !slisten {
				slisten = true
				fmt.Println(tag, "Listening on localhost:" + strconv.Itoa(port))
				fmt.Println(tag, "Enter localhost:" + strconv.Itoa(port) + "/watch.m3u8 on your player.")
				go portListen()
			}

			//Get ts'
			err = gat(tsList)
			if err != nil {
				fmt.Println(tag, "Batch unmarshal error: ", err)
			}

			if !lowA {
				//Write the final M3U8
				err = finalFileWriter(finalText)
				if err != nil {
					fmt.Println(tag, "Write final file error", err)
					return
				}
			}

			//Delete files
			err = deleteUnTs(tsList)
			if err != nil {
				fmt.Println(tag, "deleteUnTs error:", err)
			}
			fmt.Println(tag, "Ts' deleted")
		} 
	} else if bType == "watch" {
		// Download the selected M3U8
		fmt.Println(tag, "Downloading target M3U8...")
		errw, err = getM3U8(finalM3u8Url, finalM3u8Name)
		if err != nil { 
			switch errw.Error() {
				case "gm:1":
					fmt.Println(tag, "Connection error:", err)
					return
				case "gm:2":
					fmt.Println(tag, "bodyBytes error:", err)
					return
				case "gm:3":
					fmt.Println(tag, "AppScript error:", err)
					return
				case "gm:4":
					fmt.Println(tag, "Request error:", err)
					return
				case "gm:5":
					fmt.Println(tag, "Error: Its not M3U8 file!")
					return
				case "gm:6":
					fmt.Println(tag, "Write file error:", err)
					return
				default:
					fmt.Println(tag, err)
					return
			}
		}
		fmt.Println(tag, "Target M3U8 downloaded")

		// Parse TS segments
		fmt.Println(tag, "Parsing target M3U8...")
		errw, err, _, tsNum, finalText := getTsNum(finalM3u8Name)
		finalM3u8TsNum = tsNum
		if err != nil {
			switch errw.Error() {
				case "gtn:1":
					fmt.Println(tag, "Error: M3U8 not found! (gtn:1)")
					return
				case "gtn:2":
					fmt.Println(tag, "M3U8 Read Error:", err)
					return
				case "gtn:3":
					fmt.Println(tag, "Batch Create Error:", err)
					return
				default:
					fmt.Println(tag, err)
					return
			}
		}
		if tsNum <= 0 {
			fmt.Println(tag, "Error: Target M3U8 is empty!")
			return
		}
		fmt.Println(tag, strconv.Itoa(tsNum) + " ts found")

		//Write the final M3U8
		err = finalFileWriter(finalText)
		if err != nil {
			fmt.Println(tag, "Write final file error", err)
			return
		}

		fmt.Println(tag, "Listening on localhost:" + strconv.Itoa(port))
		fmt.Println(tag, "Enter localhost:" + strconv.Itoa(port) + "/watch.m3u8 on your player.")
		err = portListen()
		if err != nil {
			fmt.Println(tag, "Port listen error: ", err)
		}
	}
}

