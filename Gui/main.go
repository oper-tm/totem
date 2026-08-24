package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

type MainT struct{}
type DownloadLib struct {
	Type     string `json:"type"`
	Filename string `json:"filename"`
	Title    string `json:"title"`
	Date     string `json:"date"`
	URL      string `json:"url"`
}
type Dst struct {
	Dlib DownloadLib `json:"dlib"`
	Dtl  string      `json:"dtl"`
}

const androidPackageName = "com.totem.app"

var (
	app    *application.App
	logger *log.Logger
	ctx    context.Context
	cancel context.CancelFunc
	master bool = true
)

func init() {
	application.RegisterEvent[string]("time")
}

func main() {

	app = application.New(application.Options{
		Name:        "totem",
		Description: "Totem Relay Player",
		Services: []application.Service{
			application.NewService(&MainT{}),
		},
		Assets: application.AssetOptions{
			Handler: newAssetHandler(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Totem Player",
		Width:  1000,
		Height: 618,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(6, 7, 15),
		URL:              "/",
	})

	tet1, tet2 := os.UserConfigDir()
	fmt.Println("-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------", tet1, tet2)

	if err := InitDataPath(); err != nil {
		panic(err)
	}

	logFile, err := os.OpenFile(
		dp+"app.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		// panic(err)
	}

	logger = log.New(
		io.MultiWriter(os.Stdout, logFile),
		"",
		log.Ldate|log.Ltime,
	)

	err = loadConfig(dp + "config.json")
	if err != nil {
		// panic(err)
	}

	go logView()

	// Run the application. This blocks until the application has been exited.
	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func logView() {
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("LOG REQUEST RECEIVED")

		path := filepath.Join(dp, "app.log")
		fmt.Println("FILE:", path)

		http.ServeFile(w, r, path)
	})
	http.HandleFunc("/tsb/{path}", pRet)

	fmt.Println("Listening on :1820")
	app.Logger.Info("Listening on :1820")
	err = http.ListenAndServe(":1820", nil)
	fmt.Println(err)
}

func (d *MainT) Analyze(url string) (string, error) {
	err = loadConfig(dp + "config.json")
	if err != nil {
		panic(err)
	}

	reset()
	m3u8Url = url

	logger.Println(tag, "Analyze request: "+url)
	logger.Println(tag, "Downloading master M3U8...")

	errw, err = getM3U8(url, m3u8Name)
	if err != nil {
		switch errw.Error() {
		case "gm:1":
			logger.Println(tag, "Connection error:", err)
			return "", errw
		case "gm:2":
			logger.Println(tag, "bodyBytes error:", err)
			return "", errw
		case "gm:3":
			logger.Println(tag, "AppScript error:", err)
			return "", errw
		case "gm:4":
			logger.Println(tag, "Request error:", err)
			return "", errw
		case "gm:5":
			logger.Println(tag, "Error: Its not M3U8 file!")
			return "", errw
		case "gm:6":
			logger.Println(tag, "Write file error:", err)
			return "", errw
		default:
			logger.Println(tag, err)
			return "", errw
		}
	}
	logger.Println(tag, "Master M3U8 downloaded")

	m3u8PathUrl = strings.Split(m3u8Url, strings.Split(m3u8Url, "/")[len(strings.Split(m3u8Url, "/"))-1])[0]

	logger.Println(tag, "Checking available qualities...")

	master = true

	errw, err, qualitiesby := getQualities(m3u8Name)
	if err != nil {
		switch errw.Error() {
		case "gq:1":
			logger.Println(tag, "Error: M3U8 not found! (gq:1)")
			return "", errw
		case "gq:2":
			logger.Println(tag, "Open M3U8 file error:", err)
			return "", errw
		case "gq:3":
			logger.Println(tag, "Read file error:", err)
			return "", errw
		case "gq:4":
			logger.Println(tag, "Quality json marshal error", err)
			return "", errw
		case "gq:5":
			logger.Println(tag, "Its not a Master M3U8 file!")
			finalM3u8Url = m3u8Url
			master = false
			return finalM3u8Url, nil
		default:
			logger.Println(tag, err)
			return "", errw
		}
	}
	return string(qualitiesby), nil
}

func (d *MainT) Start(finalM3u8UrlT string, bTypeT string) (Dst, error) {
	bType = bTypeT
	ctx, cancel = context.WithCancel(context.Background())

	if master {
		if strings.Contains(finalM3u8UrlT, "http://") || strings.Contains(finalM3u8UrlT, "https://") {
			finalM3u8Url = finalM3u8UrlT
			m3u8PathUrl = strings.Split(finalM3u8Url, strings.Split(finalM3u8Url, "/")[len(strings.Split(finalM3u8Url, "/"))-1])[0]
		} else {
			finalM3u8Url = m3u8PathUrl + finalM3u8UrlT

			if strings.Contains(finalM3u8UrlT, "/") {
				m3u8PathUrlup := strings.Split(finalM3u8UrlT, "/")[len(strings.Split(finalM3u8UrlT, "/"))-1]
				m3u8PathUrlup = strings.Split(finalM3u8UrlT, m3u8PathUrlup)[0]
				m3u8PathUrl = m3u8PathUrl + m3u8PathUrlup
			}
		}
	}

	rnfname, _ := rnGenerator(10)
	outputFile = "totem" + rnfname + ".mp4"

	select {
	case <-ctx.Done():
		stopEvent()
		return Dst{}, ctx.Err()
	default:
	}

	var (
		lb_type     string = bType
		lb_filename string
		lb_title    string
		lb_date     string = time.Now().Format(time.RFC3339)
		lb_url      string = finalM3u8Url
		LBStruct    DownloadLib
	)

	dwNameFi := RandomFilename()
	lb_title = dwNameFi

	if bType == "download" {
		dwPath := filepath.Join(exportDp, dwNameFi+".mp4")
		lb_filename = dwNameFi + ".mp4"
		// Download the selected M3U8
		LBStruct = DownloadLib{
			Type:     lb_type,
			Filename: lb_filename,
			Title:    lb_title,
			Date:     lb_date,
			URL:      lb_url,
		}
		err := saveDownloadHLib(LBStruct)
		logger.Println(tag, "Downloading target M3U8...")
		errw, err = getM3U8(finalM3u8Url, finalM3u8Name)
		if err != nil {
			switch errw.Error() {
			case "gm:1":
				logger.Println(tag, "Connection error:", err)
				return Dst{}, errw
			case "gm:2":
				logger.Println(tag, "bodyBytes error:", err)
				return Dst{}, errw
			case "gm:3":
				logger.Println(tag, "AppScript error:", err)
				return Dst{}, errw
			case "gm:4":
				logger.Println(tag, "Request error:", err)
				return Dst{}, errw
			case "gm:5":
				logger.Println(tag, "Error: Its not M3U8 file!")
				return Dst{}, errw
			case "gm:6":
				logger.Println(tag, "Write file error:", err)
				return Dst{}, errw
			default:
				logger.Println(tag, err)
				return Dst{}, errw
			}
		}
		logger.Println(tag, "Target M3U8 downloaded")

		select {
		case <-ctx.Done():
			stopEvent()
			return Dst{}, ctx.Err()
		default:
		}

		// Parse TS segments
		logger.Println(tag, "Parsing target M3U8...")
		errw, err, tsList, tsNum, finalText := getTsNum(finalM3u8Name)
		finalM3u8TsNum = tsNum
		downloadedEvent(finalM3u8TsNum, true)
		if err != nil {
			switch errw.Error() {
			case "gtn:1":
				logger.Println(tag, "Error: M3U8 not found! (gtn:1)")
				return Dst{}, errw
			case "gtn:2":
				logger.Println(tag, "M3U8 Read Error:", err)
				return Dst{}, errw
			case "gtn:3":
				logger.Println(tag, "Batch Create Error:", err)
				return Dst{}, errw
			default:
				logger.Println(tag, err)
				return Dst{}, errw
			}
		}
		if tsNum <= 0 {
			logger.Println(tag, "Error: Target M3U8 is empty!")
			return Dst{}, errors.New("tmie:1")
		}
		logger.Println(tag, strconv.Itoa(tsNum)+" ts found")

		select {
		case <-ctx.Done():
			stopEvent()
			return Dst{}, ctx.Err()
		default:
		}

		//Write the final M3U8
		err = finalFileWriter(finalText)
		if err != nil {
			logger.Println(tag, "Write final file error", err)
			return Dst{}, errors.New("tmie:2")
		}

		//Get ts'
		logger.Println(string(tsList))
		_ = gat(tsList)
		select {
		case <-ctx.Done():
			stopEvent()
			return Dst{}, ctx.Err()
		default:
		}

		//Convert the ts'
		logger.Println(tag, "M3U8 is converting...")
		err = ConvertTsToMp4(tsList, dwPath)
		// err = ffmpeg(dp+"final"+finalM3u8Name, outputFile)
		if err != nil {
			logger.Println(tag, "Convert error:", err)
			logger.Println(tag, "Use the ffmpeg or other apps for convert created M3U8 file to MP4 - Filename: "+"final"+finalM3u8Name)
			return Dst{}, err
		} else {
			logger.Println(tag, "M3U8 converted")
		}

		//Delete files
		deleteTs()
		logger.Println(tag, "Ts' deleted")
	} else if bType == "stream" {
		getOk := true
		sLoad := false
		LBStruct = DownloadLib{
			Type:     lb_type,
			Filename: lb_filename,
			Title:    lb_title,
			Date:     lb_date,
			URL:      lb_url,
		}
		err := saveDownloadHLib(LBStruct)
		for getOk == true {
			logger.Println(tag, "Updating target M3U8...")
			logger.Println(tag, finalM3u8Url)
			errw, err = getM3U8(finalM3u8Url, finalM3u8Name)
			if err != nil {
				switch errw.Error() {
				case "gm:1":
					logger.Println(tag, "Connection error:", err)
					return Dst{}, errw
				case "gm:2":
					logger.Println(tag, "bodyBytes error:", err)
					return Dst{}, errw
				case "gm:3":
					logger.Println(tag, "AppScript error:", err)
					return Dst{}, errw
				case "gm:4":
					logger.Println(tag, "Request error:", err)
					return Dst{}, errw
				case "gm:5":
					logger.Println(tag, "Error: Its not M3U8 file!")
					return Dst{}, errw
				case "gm:6":
					logger.Println(tag, "Write file error:", err)
					return Dst{}, errw
				default:
					logger.Println(tag, err)
					return Dst{}, errw
				}
			}
			logger.Println(tag, "Target M3U8 updated")

			select {
			case <-ctx.Done():
				stopEvent()
				return Dst{}, ctx.Err()
			default:
			}

			// Parse TS segments
			logger.Println(tag, "Parsing target M3U8...")
			errw, err, tsList, tsNum, finalText := getTsNum(finalM3u8Name)
			finalM3u8TsNum = tsNum
			if !sLoad {
				downloadedEvent(0, true)
				sLoad = true
			}
			if err != nil {
				switch errw.Error() {
				case "gtn:1":
					logger.Println(tag, "Error: M3U8 not found! (gtn:1)")
					return Dst{}, errw
				case "gtn:2":
					logger.Println(tag, "M3U8 Read Error:", err)
					return Dst{}, errw
				case "gtn:3":
					logger.Println(tag, "Batch Create Error:", err)
					return Dst{}, errw
				case "gtn:4":
					logger.Println(tag, "Write final file error", err)
					return Dst{}, errw
				default:
					logger.Println(tag, err)
					return Dst{}, errw
				}
			}
			if tsNum <= 0 {
				logger.Println(tag, "Error: Target M3U8 is empty!")
				return Dst{}, errors.New("tmie:1")
			}

			select {
			case <-ctx.Done():
				stopEvent()
				return Dst{}, ctx.Err()
			default:
			}

			if lowA {
				//Write the final M3U8
				err = finalFileWriter(finalText)
				if err != nil {
					logger.Println(tag, "Write final file error", err)
					return Dst{}, errors.New("tmie:2")
				}
			}

			//Get ts'
			_ = gat(tsList)
			select {
			case <-ctx.Done():
				stopEvent()
				return Dst{}, ctx.Err()
			default:
			}

			if !lowA {
				//Write the final M3U8
				logger.Println("write fm ", lowA)
				err = finalFileWriter(finalText)
				if err != nil {
					logger.Println(tag, "Write final file error", err)
					return Dst{}, err
				}
			}

			//Delete files
			err = deleteUnTs(tsList)
			if err != nil {
				logger.Println(tag, "deleteUnTs error:", err)
			}
			logger.Println(tag, "Ts' deleted")
		}
	} else if bType == "watch" {
		// Download the selected M3U8
		LBStruct = DownloadLib{
			Type:     lb_type,
			Filename: lb_filename,
			Title:    lb_title,
			Date:     lb_date,
			URL:      lb_url,
		}
		err := saveDownloadHLib(LBStruct)
		logger.Println(tag, "Downloading target M3U8...")
		errw, err = getM3U8(finalM3u8Url, finalM3u8Name)
		if err != nil {
			switch errw.Error() {
			case "gm:1":
				logger.Println(tag, "Connection error:", err)
				return Dst{}, errw
			case "gm:2":
				logger.Println(tag, "bodyBytes error:", err)
				return Dst{}, errw
			case "gm:3":
				logger.Println(tag, "AppScript error:", err)
				return Dst{}, errw
			case "gm:4":
				logger.Println(tag, "Request error:", err)
				return Dst{}, errw
			case "gm:5":
				logger.Println(tag, "Error: Its not M3U8 file!")
				return Dst{}, errw
			case "gm:6":
				logger.Println(tag, "Write file error:", err)
				return Dst{}, errw
			default:
				logger.Println(tag, err)
				return Dst{}, errw
			}
		}
		logger.Println(tag, "Target M3U8 downloaded")

		select {
		case <-ctx.Done():
			stopEvent()
			return Dst{}, ctx.Err()
		default:
		}

		// Parse TS segments
		logger.Println(tag, "Parsing target M3U8...")
		errw, err, _, tsNum, finalText := getTsNum(finalM3u8Name)
		finalM3u8TsNum = tsNum
		downloadedEvent(0, true)
		if err != nil {
			switch errw.Error() {
			case "gtn:1":
				logger.Println(tag, "Error: M3U8 not found! (gtn:1)")
				return Dst{}, errw
			case "gtn:2":
				logger.Println(tag, "M3U8 Read Error:", err)
				return Dst{}, errw
			case "gtn:3":
				logger.Println(tag, "Batch Create Error:", err)
				return Dst{}, errw
			default:
				logger.Println(tag, err)
				return Dst{}, errw
			}
		}
		if tsNum <= 0 {
			logger.Println(tag, "Error: Target M3U8 is empty!")
			return Dst{}, errors.New("tmie:1")
		}
		logger.Println(tag, strconv.Itoa(tsNum)+" ts found")

		select {
		case <-ctx.Done():
			stopEvent()
			return Dst{}, ctx.Err()
		default:
		}

		//Write the final M3U8
		err = finalFileWriter(finalText)
		if err != nil {
			logger.Println(tag, "Write final file error", err)
			return Dst{}, err
		}

		select {
		case <-ctx.Done():
			stopEvent()
			return Dst{}, ctx.Err()
		default:
		}
		return Dst{Dlib: LBStruct, Dtl: "WB"}, nil
	}
	return Dst{Dlib: LBStruct, Dtl: "WC"}, nil
}

func (d *MainT) GetConfig() Config {
	return Config{
		ASA:          asa,
		BATCHSIZE:    batchSize,
		MAXSIZE:      maxSize,
		DOWNLOADMODE: downloadMode,
		LOWLATENCY:   lowA,
		WATCHPORT:    port,
		FFMPEGPATH:   ffmpegPath,
	}
}
func (d *MainT) UpdateConfig(g_asa string, g_bs, g_ms int, g_la bool) {
	asa = g_asa
	batchSize = g_bs
	maxSize = g_ms
	lowA = g_la
	if lowA {
		downloadMode = "nr"
	} else {
		downloadMode = "go"
	}

	logger.Println(asa, batchSize, maxSize, lowA)

	var config = `{
  "appScriptKey": "` + asa + `",
  "batchSize": ` + strconv.Itoa(batchSize) + `,
  "maxSize": ` + strconv.Itoa(maxSize) + `,
  "downloadMode": "` + downloadMode + `",
  "lowLatency": ` + strconv.FormatBool(lowA) + `,
  "downloadPath": "downloads/",
  "watchPort": 1819,
  "ffmpegPath": "ffmpeg.exe"
}`

	os.WriteFile(cdp, []byte(config), 0644)
}

func (d *MainT) StopAll() {
	logger.Println(tag, "Stop req...")
	cancel()
}

func (d *MainT) GetPort() int {
	return port
}

func reset() {
	finalM3u8Url = ""
	finalM3u8TsNum = 0
	batchCount = 0
	batchNum = 0
	segNum = 0
	m3u8PathUrl = ""
	m3u8Url = ""
}

func downloadedEvent(i int, count bool) {
	if count {
		logger.Println(i)
		app.Event.Emit("progressCount", i)
	} else {
		app.Event.Emit("progress", i)
	}
}
func stopEvent() {
	deleteTs()
	app.Event.Emit("stopevent")
}

func InitDataPath() error {
	if runtime.GOOS == "android" {
		dir := application.Mobile.StoragePath()
		app.Logger.Info(dir + "/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////")
		exportDir := filepath.Join(application.Mobile.StoragePath(), "exports")

		// if err := os.MkdirAll(dir, 0755); err != nil {
		// 	return errors.New("android app data dir not accessible: " + err.Error())
		// }
		if err := os.MkdirAll(exportDir, 0755); err != nil {
			return errors.New("android app data dir not accessible: " + err.Error())
		}
		dp = dir
		exportDp = exportDir
		return nil
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	dp = filepath.Join(base, "Totem") + string(os.PathSeparator)
	exportDp = filepath.Join(dp, "exports") + string(os.PathSeparator)
	os.MkdirAll(dp, 0755)
	os.MkdirAll(exportDp, 0755)
	return nil
}

// P.T
func (d *MainT) GetPluginListOU() []string {
	return getPluginList()
}

func (d *MainT) RunPluginOU(a string, b map[string]interface{}) {
	err = RunPlugin(a, b)
	if err != nil {
		logger.Println("rpe: ", err)
		return
	}
}

func (d *MainT) AddPluginOU(ctx context.Context) (string, error) {
	app := application.Get()

	srcPath, err := app.Dialog.OpenFile().
		AddFilter("JavaScript Files", "*.js").
		PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	if srcPath == "" {
		return "", nil
	}

	destPath := filepath.Join(dp, filepath.Base(srcPath))

	in, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer in.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}

	return destPath, nil
}

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomFilename() string {
	b := make([]byte, 10)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[n.Int64()]
	}
	return string(b)
}

func saveDownloadHLib(d DownloadLib) error {
	path := filepath.Join(dp, "downloadLib.json")
	var downloads []DownloadLib

	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &downloads)
	}

	downloads = append(downloads, d)

	data, err := json.MarshalIndent(downloads, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
func (d *MainT) LibLoad() ([]DownloadLib, error) {
	data, err := os.ReadFile(filepath.Join(dp, "downloadLib.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return []DownloadLib{}, nil
		}
		return nil, err
	}

	var downloads []DownloadLib
	if err := json.Unmarshal(data, &downloads); err != nil {
		return nil, err
	}

	return downloads, nil
}
func (d *MainT) BSEncoder(v string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(v))
}

// func (d *MainT) playVideo()
func GetFilesString(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	var result string

	for _, entry := range entries {
		if !entry.IsDir() {
			result += entry.Name()
		}
	}

	return result
}

func (a *MainT) OpenURL(url string, t, s bool) {
	back := "http://"
	if s {
		back = "https://"
	}
	if t {
		_ = app.Browser.OpenURL(back + url)
	} else {
		_ = app.Browser.OpenFile(back + url)
	}
}

func (a *MainT) Copy(d string) {
	_ = app.Clipboard.SetText(d)
}

func (a *MainT) TestGsi(d string) string {
	err = checkGsi(d)
	logger.Println(err)
	if err == nil {
		return "OK"
	} else {
		return err.Error()
	}
}

func (a *MainT) OpenFolder(path string) {
	fpath := path
	if path == "dpex" {
		fpath = filepath.Join(dp, "exports")
	}
	logger.Println("sdsdddddddddddddddd", fpath)
	app.Env.OpenFileManager(fpath, false)
}

func (a *MainT) DeleteTs() {
	deleteTs()
}
