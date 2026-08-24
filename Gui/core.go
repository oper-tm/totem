package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	// "fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	// "context"
	// "time"

	"embed"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/yapingcat/gomedia/go-mp4"
	"github.com/yapingcat/gomedia/go-mpeg2"
)

const (
	tag           string = "[Totem]"
	finalM3u8Name string = "target.m3u8"
	m3u8Name      string = "targetMaster.m3u8"
	charset              = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	asa          string
	dp           string
	exportDp     string
	batchSize    int
	maxSize      int
	downloadMode string
	ffmpegPath   string
	lowA         bool
	port         int
	cdp          string

	finalM3u8Url   string
	finalM3u8TsNum int

	batch SegmentBatch

	batchCount int
	batchNum   int
	segNum     int

	wg sync.WaitGroup

	err  error
	errw error

	rType string
	bType string

	qchoice int

	gqdata map[string]any

	m3u8PathUrl string
	m3u8Url     string
	outputFile  string
)

type Segment struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

type SegmentBatch struct {
	TSP [][]Segment `json:"tsp"`
}

type Config struct {
	ASA          string `json:"appScriptKey"`
	BATCHSIZE    int    `json:"batchSize"`
	MAXSIZE      int    `json:"maxSize"`
	DOWNLOADMODE string `json:"downloadMode"`
	LOWLATENCY   bool   `json:"lowLatency"`
	WATCHPORT    int    `json:"watchPort"`
	FFMPEGPATH   string `json:"ffmpegPath"`
}

var cfg Config

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

	err = os.WriteFile(dp+mname, bodyBytes, 0644)
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
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		tsTry++
		if tsTry > 1 {
			logger.Println(tag, "[ "+strconv.Itoa(tssegNum)+" / "+strconv.Itoa(finalM3u8TsNum)+" ] Retrying... (try "+strconv.Itoa(tsTry)+")")
		}
		resp, err := http.Get(
			"https://script.google.com/macros/s/" + asa + "/exec?type=1&url=" + url.QueryEscape(tsUrl),
		)

		if err != nil {
			logger.Println(tag, "[ "+strconv.Itoa(tssegNum)+" / "+strconv.Itoa(finalM3u8TsNum)+" ] Connection error:", tsName, err)
			continue
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Println(tag, "[ "+strconv.Itoa(tssegNum)+" / "+strconv.Itoa(finalM3u8TsNum)+" ] bodyBytes error:", err)
			continue
		}
		if strings.Contains(string(bodyBytes), "<!DOCTYPE html>") {
			logger.Println(tag, "[ "+strconv.Itoa(tssegNum)+" / "+strconv.Itoa(finalM3u8TsNum)+" ] AppScript error:", tsName, string(bodyBytes))
			continue
		} else if strings.Contains(string(bodyBytes), "Error:") {
			logger.Println(tag, "[ "+strconv.Itoa(tssegNum)+" / "+strconv.Itoa(finalM3u8TsNum)+" ] Request error:", tsName, string(bodyBytes))
			continue
		}

		//decode the hash

		clean := strings.Join(strings.Fields(string(bodyBytes)), "")

		rdata, err := hex.DecodeString(clean)
		if err != nil {
			logger.Println(tag, "[ "+strconv.Itoa(tssegNum)+" / "+strconv.Itoa(finalM3u8TsNum)+" ] Hex decode error:", tsName, err)
			continue
		}

		err = os.WriteFile(dp+tsName, rdata, 0644)
		if err != nil {
			logger.Println(tag, "Write file error:", tsName, err)
			continue
		}

		logger.Println(tag, "[ "+strconv.Itoa(tssegNum)+" / "+strconv.Itoa(finalM3u8TsNum)+" ] Downloaded:", tsName)
		tsOk = true
		downloadedEvent(tssegNum, false)
		return rdata
	}
	return []byte("nt")
}
func getHFile(hfurl string) []byte {
	resp, err := http.Get(
		"https://script.google.com/macros/s/" + asa + "/exec?type=1&url=" + url.QueryEscape(hfurl),
	)
	if err != nil {
		logger.Println(tag, "Connection error:", err)
		return []byte("err")
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Println(tag, "bodyBytes error:", err)
		return []byte("err")
	}
	if strings.Contains(string(bodyBytes), "<!DOCTYPE html>") {
		logger.Println(tag, "AppScript error:", string(bodyBytes))
		return []byte("err")
	} else if strings.Contains(string(bodyBytes), "Error:") {
		logger.Println(tag, "Request error:", string(bodyBytes))
		return []byte("err")
	}

	//decode the hash

	clean := strings.Join(strings.Fields(string(bodyBytes)), "")

	rdata, err := hex.DecodeString(clean)
	if err != nil {
		logger.Println(tag, "Hex decode error:", err)
		return []byte("err")
	}
	logger.Println(tag, "Downloaded:")
	return rdata
}
func getFile(murl, mname, headers string, write bool) (error, error, []byte) {
	qhs := ""
	if headers != "" {
		qhs = "&h=" + url.QueryEscape(headers)
	}
	resp, err := http.Get(
		"https://script.google.com/macros/s/" + asa + "/exec?type=0&url=" + url.QueryEscape(murl) + qhs,
	)
	if err != nil {
		return errors.New("gf:1"), err, []byte("err")
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("gf:2"), err, []byte("err")
	}

	// if strings.Contains(string(bodyBytes), "Error:") {
	// 	return errors.New("gf:3"), err, []byte("err")
	// }

	if write {
		err = os.WriteFile(dp+mname, bodyBytes, 0644)
		if err != nil {
			return errors.New("gf:4"), err, []byte("err")
		}
	}
	return nil, nil, bodyBytes
}
func getFilePost(murl, mname string, payLoad, headers map[string]interface{}, write bool) (error, error, []byte) {
	data := map[string]interface{}{
		"url":     murl,
		"headers": headers,
		"payload": payLoad,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return errors.New("gf:0"), err, []byte("err")
	}

	resp, err := http.Post(
		"https://script.google.com/macros/s/"+asa+"/exec",
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
		err = os.WriteFile(dp+mname, bodyBytes, 0644)
		if err != nil {
			return errors.New("gf:4"), err, []byte("err")
		}
	}
	return nil, nil, bodyBytes
}
func checkGsi(d string) error {
	resp, err := http.Get(
		"https://script.google.com/macros/s/" + d + "/exec?type=0&url=" + url.QueryEscape("https://http-stat.us/200"),
	)
	if err != nil {
		return errors.New("cgsi:1")
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("cgsi:1")
	}

	if string(bodyBytes) == "200 OK" {
		return nil
	} else {
		return errors.New("cgsi:2")
	}
}
func getQualities(tm3 string) (error, error, []byte) {
	if !fileExists(dp + tm3) {
		return errors.New("gq:1"), errors.New("gq:1"), []byte("err")
	}

	file, err := os.Open(dp + tm3)
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
			BANDWIDTH  string `json:"bandwidth"`
			RESOLUTION string `json:"resolution"`
			URL        string `json:"url"`
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
				exp1 = strings.Split(exp1, "#")[len(strings.Split(exp1, "#"))-1]
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
					BANDWIDTH:  bw,
					RESOLUTION: rs,
					URL:        line,
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

func getTsNum(fileName string) (error, error, []byte, int, string) {
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
		ba_result  SegmentBatch
		ba_current []Segment
		finalText  strings.Builder
		tsNum      int
		tsNum2     int
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
					if tsNum < tsNum2 || (tsNum-maxSize) > tsNum2 {
						continue
					}
				}

				if strings.Contains(tsName, "/") {
					tsName = strings.Split(tsName, "/")[len(strings.Split(tsName, "/"))-1]
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

				logger.Println(fullUrl, tsName)
				ba_current = append(ba_current, Segment{
					URL:  fullUrl,
					Name: tsName,
				})

				if bType != "stream" {
					if len(ba_current) == batchSize {
						ba_result.TSP = append(ba_result.TSP, ba_current)
						ba_current = []Segment{}
					}
				}

				if bType == "watch" || bType == "stream" {
					finalText.WriteString("/ts/" + base64.RawURLEncoding.EncodeToString([]byte(fullUrl)) + "/" + tsName)
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
	err = os.WriteFile(dp+"final"+finalM3u8Name, []byte(data), 0644)
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
		logger.Println(tag, "Batch [ "+strconv.Itoa(batchNum)+" / "+strconv.Itoa(batchCount)+" ]")
		logger.Println(tag, "[ "+strconv.Itoa(batchNum)+" / "+strconv.Itoa(batchCount)+" ] Waiting for TS downloads...'")
		for _, segment := range segments {
			logger.Println(segment.Name)
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			segNum++

			if fileExists(dp + segment.Name) {
				if bType == "download" {
					downloadedEvent(segNum, false)
				}
				logger.Println(tag, "[ "+strconv.Itoa(segNum)+" / "+strconv.Itoa(finalM3u8TsNum)+" ] This ts has already been downloaded.")
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
		logger.Println("[ " + strconv.Itoa(batchNum) + " / " + strconv.Itoa(batchCount) + " ] Ts' downloaded!")
	}
	return nil
}

func loadConfig(tcdp string) error {
	cdp = tcdp
	if !fileExists(cdp) {
		const config = `{
  "appScriptKey": "",
  "batchSize": 5,
  "maxSize": 10,
  "downloadMode": "go",
  "lowLatency": false,
  "downloadPath": "downloads/",
  "watchPort": 1819,
  "ffmpegPath": "ffmpeg.exe"
}`

		os.WriteFile(cdp, []byte(config), 0644)
	}
	data, err := os.ReadFile(cdp)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	asa = cfg.ASA
	batchSize = cfg.BATCHSIZE
	maxSize = cfg.MAXSIZE
	downloadMode = cfg.DOWNLOADMODE
	lowA = cfg.LOWLATENCY
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
	logger.Println(tag, "if its not worked use this:", ffmpegPath, "-i", fileName, "-c", "copy", outputName)
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

func rnGenerator(length int) (string, error) {
	b := make([]byte, length)
	r := make([]byte, length)

	if _, err := rand.Read(r); err != nil {
		return "", err
	}

	for i := range b {
		b[i] = charset[int(r[i])%len(charset)]
	}

	return string(b), nil
}

func newAssetHandler(assets embed.FS) http.Handler {
	fileServer := application.AssetFileServerFS(assets)

	streamMux := http.NewServeMux()
	streamMux.HandleFunc("/ts/{url}/{name}", tsHandler)
	streamMux.HandleFunc("/watch", m3u8Handler)
	streamMux.HandleFunc("/photo/{url}", ptph)
	streamMux.HandleFunc("/pret/{path}", pRet)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		path := r.URL.Path
		if strings.HasPrefix(path, "/pret/") {
			w.Header().Set("Content-Type", "video/mp4")
		}
		switch {
		case strings.HasPrefix(path, "/ts/"),
			strings.HasPrefix(path, "/pret/"),
			path == "/watch",
			strings.HasPrefix(path, "/photo/"):
			streamMux.ServeHTTP(w, r)
		default:
			fileServer.ServeHTTP(w, r)
		}
	})
}
func tsHandler(w http.ResponseWriter, r *http.Request) {
	logger.Println(tag, "tsHandler CALLED for:", r.URL.Path)
	select {
	case <-ctx.Done():
		stopEvent()
		return
	default:
	}
	if bType == "watch" {
		urlE := r.PathValue("url")
		name := r.PathValue("name")
		urlD, err := base64.RawURLEncoding.DecodeString(urlE)
		url := string(urlD)
		if err != nil {
			logger.Println(tag, "BASE64 Url Decode on tsHandler:", err)
			return
		}

		logger.Println(tag, "Listen "+name)

		wb := getTs(url, name, 0, nil)

		w.Write(wb)
	} else {
		name := dp + r.PathValue("name")
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
func ptph(w http.ResponseWriter, r *http.Request) {
	url2 := r.PathValue("url")
	urlD, _ := base64.RawURLEncoding.DecodeString(url2)
	url := string(urlD)

	logger.Println(tag, "Listen "+url)

	wb := getHFile(url)

	w.Write(wb)
}
func pRet(w http.ResponseWriter, r *http.Request) {
	tpath := r.PathValue("path")

	pathD, err := base64.RawURLEncoding.DecodeString(tpath)
	if err != nil {
		http.Error(w, "bad path", 400)
		return
	}

	path := filepath.Join(dp, string(pathD))

	logger.Println(path)

	file, err := os.Open(path)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func ConvertTsToMp4(tsFiles []byte, outputPath string) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	muxer, err := mp4.CreateMp4Muxer(outFile)
	if err != nil {
		return err
	}

	var batch SegmentBatch
	if err := json.Unmarshal(tsFiles, &batch); err != nil {
		return err
	}

	var videoTrack, audioTrack uint32
	var hasVideo, hasAudio bool
	var ptsOffset, dtsOffset uint64
	var lastPts, lastDts uint64
	var isFirstSegment = true

	for _, segments := range batch.TSP {
		for _, segment := range segments {
			f, err := os.Open(dp + segment.Name)
			if err != nil {
				return err
			}

			var sawFirstFrame bool

			demuxer := mpeg2.NewTSDemuxer()
			demuxer.OnFrame = func(cid mpeg2.TS_STREAM_TYPE, frame []byte, pts uint64, dts uint64) {
				if !sawFirstFrame {
					sawFirstFrame = true
					if isFirstSegment || pts > lastPts {
						// pts از قبل پیوسته‌ست (مطلقه)، نیازی به offset نیست
						ptsOffset = 0
						dtsOffset = 0
					} else {
						// pts این segment ریست شده، بچسبونش به انتهای قبلی
						ptsOffset = lastPts
						dtsOffset = lastDts
					}
					logger.Printf("[%s] raw first pts=%d dts=%d -> offset pts=%d dts=%d\n",
						segment.Name, pts, dts, ptsOffset, dtsOffset)
				}

				adjPts := pts + ptsOffset
				adjDts := dts + dtsOffset

				switch cid {
				case mpeg2.TS_STREAM_H264:
					if !hasVideo {
						videoTrack = muxer.AddVideoTrack(mp4.MP4_CODEC_H264)
						hasVideo = true
					}
					muxer.Write(videoTrack, frame, adjPts, adjDts)
				case mpeg2.TS_STREAM_AAC:
					if !hasAudio {
						audioTrack = muxer.AddAudioTrack(mp4.MP4_CODEC_AAC)
						hasAudio = true
					}
					muxer.Write(audioTrack, frame, adjPts, adjDts)
				}
				lastPts, lastDts = adjPts, adjDts
			}

			if err := demuxer.Input(f); err != nil {
				f.Close()
				return err
			}
			f.Close()

			isFirstSegment = false
		}
	}

	return muxer.WriteTrailer()
}
