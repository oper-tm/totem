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
	"bytes"
)

var (
	asa string
	dp string
	batchSize int
	maxSize int
	downloadMode string
	ffmpegPath string
	lowA bool
	port int

	finalM3u8Url string
	finalM3u8TsNum int

	batch SegmentBatch

	batchCount int
	batchNum int
	segNum int

	wg sync.WaitGroup
)

type Segment struct {
	URL string `json:"url"`
	Name string `json:"name"`
}

type SegmentBatch struct {
	TSP [][]Segment `json:"tsp"`
}

func getM3U8(murl, mname string) (error, error) {
	resp, err := http.Get(
	"https://script.google.com/macros/s/" + asa + "/exec?type=0&url=" + url.QueryEscape(murl),
	)
	if err != nil {
		return errors.New("gm:1"), err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("gm:2"), err
	}

	if strings.Contains(string(bodyBytes), "<!DOCTYPE html>") {
		return errors.New("gm:3"), errors.New("Html codes")
	} else if strings.Contains(string(bodyBytes), "Error:") {
		return errors.New("gm:4"), errors.New(string(bodyBytes))
	} else if !strings.HasPrefix(string(bodyBytes), "#EXTM3U") {
		return errors.New("gm:5"), errors.New("gm:5")
	}

	err = os.WriteFile(dp + mname, bodyBytes, 0644)
	if err != nil {
		return errors.New("gm:6"), err
	}
	return nil, nil
}

func getTs(tsUrl, tsName string, tssegNum int, wg *sync.WaitGroup) []byte {
	if downloadMode == "go" && bType != "watch" {
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
		return rdata
	}
	return []byte("nt")
}
func getFile(murl, mname string, write bool) (error, error, []byte) {
	resp, err := http.Get(
	"https://script.google.com/macros/s/" + asa + "/exec?type=0&url=" + url.QueryEscape(murl),
	)
	if err != nil {
		return errors.New("gf:1"), err, []byte("err")
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("gf:2"), err, []byte("err")
	}

	if strings.Contains(string(bodyBytes), "Error:") {
		return errors.New("gf:3"), err, []byte("err")
	}

	if write {
		err = os.WriteFile(dp + mname, bodyBytes, 0644)
		if err != nil {
			return errors.New("gf:4"), err, []byte("err")
		}
	}
	return nil, nil, bodyBytes
}
func getFilePost(murl, mname string, payLoad, headers map[string]interface{}, write bool) (error, error, []byte) {
	data := map[string]interface{}{
		"url": murl,
		"headers": headers,
		"payload": payLoad,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return errors.New("gf:0"), err, []byte("err")
	}

	resp, err := http.Post(
		"https://script.google.com/macros/s/" + asa + "/exec",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return errors.New("gf:1"), err, []byte("err")
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("gf:2"), err, []byte("err")
	}

	if strings.Contains(string(bodyBytes), "Error:") {
		return errors.New("gf:3"), err, []byte("err")
	}

	if write {
		err = os.WriteFile(dp + mname, bodyBytes, 0644)
		if err != nil {
			return errors.New("gf:4"), err, []byte("err")
		}
	}
	return nil, nil, bodyBytes
}
func getQualities() (error, error, []byte) {
	if !fileExists(dp + m3u8Name) {
		return errors.New("gq:1"), errors.New("gq:1"), []byte("err")
	}

	file, err := os.Open(dp + m3u8Name)
	if err != nil {
		return errors.New("gq:2"), err, []byte("err")
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return errors.New("gq:3"), err, []byte("err")
	}

	content := string(data)

	file.Seek(0, 0)

	scanner := bufio.NewScanner(file)
	
	if strings.Contains(content, "#EXT-X-STREAM-INF") || strings.Contains(content, "#EXT-X-I-FRAME-STREAM-INF") {
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
			return errors.New("gq:4"), err, []byte("err")
		}
		return nil, nil, sq
	} else {
		return errors.New("gq:5"), errors.New("gq:nm"), []byte("err")
	}
}

func getTsNum(fileName string) (error, error, []byte, int, string){
	if !fileExists(dp + m3u8Name) {
		return errors.New("gtn:1"), errors.New("gtn:1"), []byte("err"), 0, "err"
	}

	file, err := os.Open(dp + fileName)
	if err != nil {
		return errors.New("gtn:2"), err, []byte("err"), 0, "err"
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)

	var (
		ba_result SegmentBatch
		ba_current []Segment
		finalText strings.Builder
		tsNum int
		tsNum2 int
	)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
				
		if strings.Contains(line, ".ts?") || strings.HasSuffix(line, ".ts") {
			tsNum++
		}
	}

	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			finalText.WriteString(line)
			finalText.WriteByte('\n')
		} else {
			if strings.Contains(line, ".ts?") || strings.HasSuffix(line, ".ts") {
				var tsName string = line
				var fullUrl string

				tsNum2++

				if bType == "stream" {
					if tsNum < tsNum2 || (tsNum - maxSize) > tsNum2 {
						continue
					}
				}

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

				if bType != "stream" {
					if len(ba_current) == batchSize {
						ba_result.TSP = append(ba_result.TSP, ba_current)
						ba_current = []Segment{}
					}
				}
				
				if bType == "watch" || bType == "stream" {
					finalText.WriteString("http://localhost:" + strconv.Itoa(port) + "/ts?url=" + url.QueryEscape(fullUrl) + "&name=" + tsName)
				} else {
					finalText.WriteString(tsName)
				}
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
		return errors.New("gtn:3"), err, []byte("err"), 0, "err"
	}

	return nil, nil, data, tsNum, finalText.String()
}

func finalFileWriter(data string) error {
	err = os.WriteFile(dp + "final" + finalM3u8Name, []byte(data), 0644)
	if err != nil {
		return err
	}
	return nil
}

func gat(tsList []byte) error {
	err = json.Unmarshal(tsList, &batch)
	if err != nil {
		return err
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
				_ = getTs(segment.URL, segment.Name, segNum, &wg)
			} 
		}
		wg.Wait()
		fmt.Println("[ " + strconv.Itoa(batchNum) + " / " + strconv.Itoa(batchCount) + " ] Ts' downloaded!")
	}
	return nil
}

func portListen() error {
	http.HandleFunc("/ts", tsHandler)
	http.HandleFunc("/watch.m3u8", m3u8Handler)

	err := http.ListenAndServe(":" + strconv.Itoa(port), nil)
	if err != nil {
		return err
	}
	return nil
}

func tsHandler(w http.ResponseWriter, r *http.Request) {
	if bType == "watch" {
		url := r.URL.Query().Get("url")
		name := r.URL.Query().Get("name")

		fmt.Println(tag, "Listen " + name)

		wb := getTs(url, name, 0, nil)

		w.Write(wb)
	} else {
		name := dp + r.URL.Query().Get("name")
		file, err := os.Open(name)
		if err != nil {
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)

		w.Write(data)
	}
}
func m3u8Handler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(dp + "finaltarget.m3u8")
	if err != nil {
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)

	w.Write(data)
}

func loadConfig() error {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return err
	}

	var cfg struct {
		ASA string `json:"appScriptKey"`
		BATCHSIZE int `json:"batchSize"`
		MAXSIZE int `json:"maxSize"`
		DOWNLOADMODE string `json:"downloadMode"`
		LOWLATENCY bool `json:"lowLatency"`
		DOWNLOADPATH string `json:"downloadPath"`
		WATCHPORT int `json:"watchPort"`
		FFMPEGPATH string `json:"ffmpegPath"`
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	asa = cfg.ASA
	batchSize = cfg.BATCHSIZE
	maxSize = cfg.MAXSIZE
	downloadMode = cfg.DOWNLOADMODE
	lowA = cfg.LOWLATENCY
	dp = cfg.DOWNLOADPATH
	port = cfg.WATCHPORT
	ffmpegPath = cfg.FFMPEGPATH
	
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	} else {
		ffmpegPath = normalizeExecutablePath(cfg.FFMPEGPATH)
	}

	return nil
}

func downloadDirectoryCheck() error {
	if _, err := os.Stat(dp); os.IsNotExist(err) {
		err = os.MkdirAll(dp, 0755)
		return err
	}
	return nil
}

func ffmpeg(fileName, outputName string) error {
	cmd := exec.Command(
		ffmpegPath,
		"-i", fileName,
		"-c", "copy",
		outputName,
	)

	_, err := cmd.CombinedOutput()
	fmt.Println(tag, "if its not worked use this:", ffmpegPath, "-i", fileName, "-c", "copy", outputName)
	if err != nil {
		return err
	}

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

func deleteTs() {
	files, _ := os.ReadDir(dp)
	for _, f := range files {
		name := f.Name()

		if strings.HasSuffix(name, ".ts") {
			os.Remove(dp + name)
		}
	}
}

func deleteUnTs(tsList []byte) error {
	err = json.Unmarshal(tsList, &batch)
	if err != nil {
		return err
	}

	exists := make(map[string]bool)

	for _, segments := range batch.TSP {
		for _, segment := range segments {
			exists[segment.Name] = true
		}
	}

	files, _ := os.ReadDir(dp)
	for _, f := range files {
		name := f.Name()

		if strings.HasSuffix(name, ".ts") {
			if !exists[name] {
				os.Remove(dp + name)
			}
		}
	}

	return nil
}