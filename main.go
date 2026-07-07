package main

import (
	"fmt"
	"unicode/utf8"
	"os"
	"os/exec"
	"strings"
	"encoding/json"
	"encoding/hex"
	"io"
	"net/http"
    "net/url"
	"errors"
	"bufio"
	"strconv"
	"sync"
	"path/filepath"
)

const (
	tag string = "[Mard]"
)

var (
	err error

	asa string
	dp string
	batchSize int
	downloadMode string
	ffmpegPath string

	m3u8PathUrl string
	m3u8Url string
	m3u8Name string = "targetMaster.m3u8"
	outputFile string

	finalM3u8Url string
	finalM3u8Name string = "target.m3u8"
	finalM3u8TsNum int

	batch SegmentBatch

	batchCount int
	batchNum int
	segNum int

	gqdata map[string]any
	qchoice int

	wg sync.WaitGroup
)

type Segment struct {
	URL string `json:"url"`
	Name string `json:"name"`
}

type SegmentBatch struct {
	TSP [][]Segment `json:"tsp"`
}

func main() {
	fmt.Println(tag, "Made by Totem - M.A.R.D 1.0.0 - 2026")

	if len(os.Args) < 2 {
		fmt.Println(tag, "Usage: masm <m3u8_url> <output_file>")
		return
	}

	// Get the M3U8 URL
	m3u8Url = os.Args[1]
	outputFile = os.Args[2]

	// Load config
	fmt.Println(tag, "Loading config...")
	err = loadConfig()
	if err != nil {
		fmt.Println(tag, "Config Error: ", err)
		return
	} else if asa == "EMP" {
		fmt.Println(tag, "Config Error: AppScript api code is missing")
	}

	// Check download directory
	if _, err := os.Stat(dp); os.IsNotExist(err) {
		err = os.MkdirAll(dp, 0755)
		if err != nil {
			fmt.Println(tag, "Create download folder error:", err)
			return
		}
	}

	// Download the master M3U8
	fmt.Println(tag, "Downloading master M3U8...")
	err = getM3U8(m3u8Url, m3u8Name)
	if err != nil { return }
	fmt.Println(tag, "Master M3U8 downloaded")

	// Set the base M3U8 URL
	m3u8PathUrl = strings.Split(m3u8Url, strings.Split(m3u8Url, "/")[len(strings.Split(m3u8Url, "/")) - 1])[0]

	// Detect available qualities
	fmt.Println(tag, "Checking available qualities...")
	err, qualitiesby := getQualities()
	if err != nil { 
		if err != errors.New("gq:nm") {
			finalM3u8Url = m3u8Url
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

		finalM3u8Url = m3u8PathUrl + selectedQuality["url"].(string)

		if strings.Contains(selectedQuality["url"].(string), "/") {
			m3u8PathUrlup := strings.Split(selectedQuality["url"].(string), "/")[len(strings.Split(selectedQuality["url"].(string), "/")) - 1]
			m3u8PathUrlup = strings.Split(selectedQuality["url"].(string), m3u8PathUrlup)[0]
			m3u8PathUrl = m3u8PathUrl + m3u8PathUrlup
		}
	}

	// Download the selected M3U8
	fmt.Println(tag, "Downloading target M3U8...")
	err = getM3U8(finalM3u8Url, finalM3u8Name)
	if err != nil { return }
	fmt.Println(tag, "Target M3U8 downloaded")

	// Parse TS segments
	fmt.Println(tag, "Parsing target M3U8...")
	err, tsList, tsNum := getTsNum(finalM3u8Name)
	finalM3u8TsNum = tsNum
	if err != nil { return }
	if tsNum <= 0 {
		fmt.Println(tag, "Error: Target M3U8 is empty!")
		return
	}
	fmt.Println(tag, strconv.Itoa(tsNum) + " ts found")
	
	//Get ts'

	err = json.Unmarshal(tsList, &batch)
	if err != nil {
		fmt.Println(tag, "Batch unmarshal error: ", err)
		return
	}

	batchCount = len(batch.TSP)

	for _, segments := range batch.TSP {
		batchNum++
		fmt.Println(tag, "Batch [ " + strconv.Itoa(batchNum) + " / " + strconv.Itoa(batchCount) + " ]")
		fmt.Println(tag, "[ " + strconv.Itoa(batchNum) + " / " + strconv.Itoa(batchCount) + " ] Waiting for TS downloads...'")
		for _, segment := range segments {
			segNum++

			if fileExists(dp + segment.Name) {
				fmt.Println(tag, "[ " + strconv.Itoa(segNum) + " / " + strconv.Itoa(finalM3u8TsNum) + " ] This ts has already been downloaded.")
				continue
			}

			if downloadMode == "go" {
				wg.Add(1)
				go getTs(segment.URL, segment.Name, segNum, &wg)
			} else {
				getTs(segment.URL, segment.Name, segNum, &wg)
			} 
		}
		wg.Wait()
		fmt.Println("[ " + strconv.Itoa(batchNum) + " / " + strconv.Itoa(batchCount) + " ] Ts' downloaded!")
	}

	//Convert the ts'
	fmt.Println(tag, "M3U8 is converting...")
	err = ffmpeg(dp + "final" + finalM3u8Name, outputFile)
	if err != nil {
		fmt.Println(tag, "Use the ffmpeg or other apps for convert created M3U8 file to MP4 - Filename: " + "final" + finalM3u8Name)
		return
	}
	fmt.Println(tag, "M3U8 converted")

	//Delete files
	files, _ := os.ReadDir(dp)
	for _, f := range files {
		name := f.Name()

		if strings.HasSuffix(name, ".ts") {
			os.Remove(dp + name)
		}
	}

	fmt.Println(tag, "Ts' deleted")
}

func getM3U8(murl, mname string) error {
	resp, err := http.Get(
	"https://script.google.com/macros/s/" + asa + "/exec?type=0&url=" + url.QueryEscape(murl),
	)
	if err != nil {
		fmt.Println(tag, "Connection error:", mname, err)
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(tag, "bodyBytes error:", mname, err)
		return err
	}

	if strings.Contains(string(bodyBytes), "<!DOCTYPE html>") {
		fmt.Println(tag, "AppScript error:", mname, string(bodyBytes))
		return errors.New("gm:1")
	} else if strings.Contains(string(bodyBytes), "Error:") {
		fmt.Println(tag, "Request error:", mname, string(bodyBytes))
		return errors.New("gm:2")
	} else if !strings.HasPrefix(string(bodyBytes), "#EXTM3U") {
		fmt.Println(tag, "Error: Its not M3U8 file!")
		return errors.New("gm:3")
	}

	err = os.WriteFile(dp + mname, bodyBytes, 0644)
	if err != nil {
		fmt.Println(tag, "Write file error:", mname, err)
		return err
	}
	return nil
}

func getTs(tsUrl, tsName string, tssegNum int, wg *sync.WaitGroup) {
	if downloadMode == "go" {
		defer wg.Done()
	}
	tsOk := false
	var tsTry int

	for !tsOk {
		tsTry++
		if tsTry > 1 {
			fmt.Println(tag, "[ " + strconv.Itoa(tssegNum) + " / " + strconv.Itoa(finalM3u8TsNum) + " ] Retrying... (try " + strconv.Itoa(tsTry) + ")")
		}
		resp, err := http.Get(
		"https://script.google.com/macros/s/" + asa + "/exec?type=1&url=" + url.QueryEscape(tsUrl),
		)
		if err != nil {
			fmt.Println(tag, "[ " + strconv.Itoa(tssegNum) + " / " + strconv.Itoa(finalM3u8TsNum) + " ] Connection error:", tsName, err)
			continue
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(tag, "[ " + strconv.Itoa(tssegNum) + " / " + strconv.Itoa(finalM3u8TsNum) + " ] bodyBytes error:" ,err)
			continue
		}

		if strings.Contains(string(bodyBytes), "<!DOCTYPE html>") {
			fmt.Println(tag, "[ " + strconv.Itoa(tssegNum) + " / " + strconv.Itoa(finalM3u8TsNum) + " ] AppScript error:", tsName, string(bodyBytes))
			continue
		} else if strings.Contains(string(bodyBytes), "Error:") {
			fmt.Println(tag, "[ " + strconv.Itoa(tssegNum) + " / " + strconv.Itoa(finalM3u8TsNum) + " ] Request error:", tsName, string(bodyBytes))
			continue
		}

		//decode the hash

		clean := strings.Join(strings.Fields(string(bodyBytes)), "")

		rdata, err := hex.DecodeString(clean)
		if err != nil {
			fmt.Println(tag, "[ " + strconv.Itoa(tssegNum) + " / " + strconv.Itoa(finalM3u8TsNum) + " ] Hex decode error:", tsName, err)
			continue
		}


		err = os.WriteFile(dp + tsName, rdata, 0644)
		if err != nil {
			fmt.Println(tag, "Write file error:", tsName, err)
			continue
		}

		fmt.Println(tag, "[ " + strconv.Itoa(tssegNum) + " / " + strconv.Itoa(finalM3u8TsNum) + " ] Downloaded:", tsName)
		tsOk = true
	}
}

func getQualities() (error, []byte) {
	if !fileExists(dp + m3u8Name) {
		fmt.Println(tag, "Error: M3U8 not found! (gq:1)")
		return errors.New("gq:1"), []byte("err")
	}

	file, err := os.Open(dp + m3u8Name)
	if err != nil {
		fmt.Println(tag, "Open M3U8 file error:", err)
		return err, []byte("err")
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(tag, "Read file error:", err)
		return err, []byte("err")
	}

	content := string(data)

	file.Seek(0, 0)

	scanner := bufio.NewScanner(file)
	
	if strings.Contains(content, "#EXT-X-STREAM-INF") || strings.Contains(content, "#EXT-X-MEDIA") || strings.Contains(content, "#EXT-X-I-FRAME-STREAM-INF") {
		type QualitySF struct {
			BANDWIDTH string `json:"bandwidth"`
			RESOLUTION string `json:"resolution"`
			URL string `json:"url"`
		}

		type finalRS struct {
			Qualities []QualitySF `json:"qualities"`
		}

		qualitiesJson := finalRS{}

		for scanner.Scan() {
			var (
				bw string = "null"
				rs string = "null"
			)

			line := strings.TrimSpace(scanner.Text())

			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			if strings.Contains(line, ".m3u8?") || strings.HasSuffix(line, ".m3u8") {
				exp1 := strings.Split(content, line)[0]
				exp1 = strings.Split(exp1, "#")[len(strings.Split(exp1, "#")) - 1]
				exp1 = strings.ReplaceAll(exp1, "\n", "")
				exp1 = strings.ReplaceAll(exp1, "\r", "")

				if strings.Contains(exp1, "BANDWIDTH") {
					bw = strings.Split(exp1, "BANDWIDTH=")[1]
					if strings.Contains(bw, ",") {
						bw = strings.Split(bw, ",")[0]
					}
				}

				if strings.Contains(exp1, "RESOLUTION") {
					rs = strings.Split(exp1, "RESOLUTION=")[1]
					if strings.Contains(rs, ",") {
						rs = strings.Split(rs, ",")[0]
					}
				}

				qualitiesJson.Qualities = append(qualitiesJson.Qualities, QualitySF{
					BANDWIDTH: bw,
					RESOLUTION: rs,
					URL: line,
				})
			}
		}

		sq, err := json.Marshal(qualitiesJson)
		if err != nil {
			fmt.Println(tag, "Quality json marshal error", err)
			return err, []byte("err")
		}
		return nil, sq
	} else {
		fmt.Println(tag, "Its not a Master M3U8 file!")
		return errors.New("gq:nm"), []byte("err")
	}
}

func getTsNum(fileName string) (error, []byte, int){
	if !fileExists(dp + m3u8Name) {
		fmt.Println(tag, "Error: M3U8 not found! (gtn:1)")
		return errors.New("gtn:1"), []byte("err"), 0
	}

	file, err := os.Open(dp + fileName)
	if err != nil {
		fmt.Println(tag, "M3U8 Read Error:", dp + m3u8Name, err)
		return err, []byte("err"), 0
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)

	var (
		ba_result SegmentBatch
		ba_current []Segment
		finalText strings.Builder
		tsNum int
	)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			finalText.WriteString(line)
			finalText.WriteByte('\n')
		} else {
			if strings.Contains(line, ".ts?") || strings.HasSuffix(line, ".ts") {
				var tsName string = line
				var fullUrl string
				tsNum++
				if strings.Contains(tsName, "/") {
					tsName = strings.Split(tsName, "/")[len(strings.Split(tsName, "/")) - 1]
				}

				if strings.Contains(tsName, "?") {
					tsName = strings.Split(tsName, "?")[0]
				}

				tsNameForLenCheck := strings.Split(tsName, ".ts")[0]
				if len(tsNameForLenCheck) > 150 {
					b := []byte(tsNameForLenCheck[:150])
					for !utf8.Valid(b) {
						b = b[:len(b)-1]
					}
					tsNameForLenCheck = string(b)
					tsName = tsNameForLenCheck + ".ts"
				}

				if strings.Contains(line, "https://") || strings.Contains(line, "http://") {
					fullUrl = line
				} else {
					fullUrl = m3u8PathUrl + line
				}

				ba_current = append(ba_current, Segment{
					URL: fullUrl,
					Name: tsName,
				})

				if len(ba_current) == batchSize {
					ba_result.TSP = append(ba_result.TSP, ba_current)
					ba_current = []Segment{}
				}

				finalText.WriteString(tsName)
				finalText.WriteByte('\n')
			} else {
				finalText.WriteString(line)
				finalText.WriteByte('\n')
			}
		}
	}
	if len(ba_current) > 0 {
		ba_result.TSP = append(ba_result.TSP, ba_current)
	}

	data, err := json.Marshal(ba_result)
	if err != nil {
		fmt.Println(tag, "Batch Create Error:", dp + m3u8Name, err)
		return err, []byte("err"), 0
	}

	err = os.WriteFile(dp + "final" + finalM3u8Name, []byte(finalText.String()), 0644)
	if err != nil {
		fmt.Println(tag, "Write final file error", err)
		return err, []byte("err"), 0
	}

	return nil, data, tsNum
}

func ffmpeg(fileName, outputName string) error {
	cmd := exec.Command(
		ffmpegPath,
		"-i", fileName,
		"-c", "copy",
		outputName,
	)

	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(tag, "Convert error:", err)
		return err
	}

	return nil
}

func loadConfig() error {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return err
	}

	var cfg struct {
		ASA string `json:"asa"`
		BATCHSIZE int `json:"batchSize"`
		DOWNLOADMODE string `json:"downloadMode"`
		DOWNLOADPATH string `json:"downloadPath"`
		FFMPEGPATH string `json:"ffmpegPath"`
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	asa = cfg.ASA
	batchSize = cfg.BATCHSIZE
	downloadMode = cfg.DOWNLOADMODE
	dp = cfg.DOWNLOADPATH
	ffmpegPath = normalizeExecutablePath(cfg.FFMPEGPATH)

	return nil
}

func normalizeExecutablePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, ".\\") {
		return path
	}

	return "." + string(filepath.Separator) + path
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}