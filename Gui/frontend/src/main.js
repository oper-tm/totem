import { Analyze } from "../bindings/changeme/maint"
import { Start } from "../bindings/changeme/maint"
import { StopAll } from "../bindings/changeme/maint"
import { GetPort } from "../bindings/changeme/maint"
import { GetConfig } from "../bindings/changeme/maint"
import { UpdateConfig } from "../bindings/changeme/maint"
import { RunPluginOU } from "../bindings/changeme/maint"
import { GetPluginListOU } from "../bindings/changeme/maint"
import { AddPluginOU } from "../bindings/changeme/maint"
import { LibLoad } from "../bindings/changeme/maint"
import { BSEncoder } from "../bindings/changeme/maint"
import { OpenURL } from "../bindings/changeme/maint"
import { Copy } from "../bindings/changeme/maint"
import { TestGsi } from "../bindings/changeme/maint"
import { OpenFolder } from "../bindings/changeme/maint"
import { DeleteTs } from "../bindings/changeme/maint"
import { Events } from '@wailsio/runtime';
import Hls from "hls.js";
import Plyr from 'plyr';
import 'plyr/dist/plyr.css';

const platform = (() => {
  if (typeof window.wails?.platform === 'function') return window.wails.platform(); // Android
  if (window.webkit?.messageHandlers?.external) return 'ios';
  return 'desktop';
})();

const isIOS     = platform === 'ios';
const isAndroid = platform === 'android';
const isMobile  = isIOS || isAndroid;

var port = 1819;
var boot = document.getElementById("boot");
var a_btn = document.getElementById("a_btn");
var a_inp = document.getElementById("u_inp");
var navBack = document.getElementById("navback");
var watchNavBtnB = document.getElementById("watchNavBtnB");
var watchNavBtn = document.getElementById("watchNavBtn");
var qualitySelectDiv = document.getElementById("qualitySelect");
var modeSelectDiv = document.getElementById("modeSelect");
var statusDiv = document.getElementById("status");
var watchVideo = new Plyr('#video');
var elmvideo = document.getElementById("video");
var errorbox = document.getElementById("error")
var progressDiv;
var progressData;

var alPlSl = 0;
var alPlSln = "normal";

var PMode;

var progressCount = 0;
var progress = 0;
var progressPer = 0;

var s_gsa = document.getElementById("s_gsa");
var s_bs = document.getElementById("s_bs");
var s_ms = document.getElementById("s_ms");
var s_la_on = document.getElementById("s_la_on");
var s_la_off = document.getElementById("s_la_off");
var saveChangesBtn = document.getElementById("saveChangesBtn");

var lp = document.getElementById("lp");

var Cnow = "main";

var serverRun = false;

var N_gsa;
var N_bs;
var N_ms;
var N_la;

var gsa;
var bs;
var ms;
var la;

var Cmain = document.getElementById("main");
var Cwatch = document.getElementById("watch");
var Csetting = document.getElementById("setting");
var Cext = document.getElementById("ext");
var Clib = document.getElementById("lib");

var extContent = document.getElementById("extContent");
var extList = document.getElementById("extList");
var addPluginBtn = document.getElementById("addPluginBtn");
var MextList = document.getElementById("extList_m");
var MaddPluginBtn = document.getElementById("addPluginBtn_m");
var smaext = document.getElementById("smaext");

var setupback = document.getElementById("setupback");

var testGsiInp = document.getElementById("t_gsa");
var gsiStatusTxt = document.getElementById("gsiStatusTxt");

var dwpathboxw = document.getElementById("dwpathboxw");
var dwpathboxm = document.getElementById("dwpathboxm");
var m_dwpb_filename = document.getElementById("m_dwpb_filename");
var m_dwpb_code = document.getElementById("m_dwpb_code");
var m_dwpb_copy = document.getElementById("m_dwpb_copy");
var m_dwpb_opbtn = document.getElementById("m_dwpb_opbtn");
var w_dwpb_filename = document.getElementById("w_dwpb_filename");
var w_dwpb_code = document.getElementById("w_dwpb_code");
var w_dwpb_copy = document.getElementById("w_dwpb_copy");

var notRun = document.getElementById("notRun");

const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));

setInterval(serverRunCheck, 1000);
await getConfig();
console.log("ssssssssss-" + gsa + "-ssssssssssss");
if (gsa == "") {
    setupback.style.display = "flex";
    setupAsgPage(0);
}
setTimeout(async () => {
    await slplselect(alPlSl, alPlSln);
}, 100)
setTimeout(() => {
    boot.style.display = "none";
}, 2000);
Events.On("progressCount", (value) => {
    progressCount = value.data;
    progress = 0;
    progressPer = 0;
    progressSetup();
})

Events.On("progress", (value) => {
    progress++
    progressUpdate()
    console.log(progress);
})

Events.On("stopevent", () => {
    a_btn.disabled = false;
    clearMain();
    abtn(true);
})

Events.On("ih", (value) => {
    extContent.insertAdjacentHTML("beforeend", value.data);
})

Events.On("ihrem", () => {
    extContent.innerHTML = "";
})

Events.On("fuinp", (value) => {
    fuinp(value.data);
})

Events.On("lp", (value) => {
    if (value.data == "t") {
        lp.style.display = "flex";
    } else {
        lp.style.display = "none";
    }
})

Events.On("pt", (value) => {
    console.log(value)
    Cext.insertAdjacentHTML("beforeend", ` <!-- P.T Dont Share -->
    <div class="ptitem" onclick="pton('${value.data.href}')">
        <img src="/photo/${value.data.thumb}" alt="" style="width: 100%;">
        <a style="color: white;">${value.data.title}</a>
    </div><br>`);
})

window.show = (c) => show(c);
window.nChange = () => nChange();
window.saveChanges = () => updateConfig();
window.hideSMExt = () => hideSMExt();
window.selectPluginStyle = (d) => selectPluginStyle(d);
window.fuinp = (d) => fuinp(d);
window.watchNavCrt = () => watchNavCrt();
window.slplsel = (a, b) => slplselect(a, b);
window.copyCode = (a) => copy(a);
window.copy = (a) => Copy(a);
window.openDownloadLocalMenu = (a, b) => openDownloadLocalMenu(a, b);
window.closeDownloadLocalMenu = () => closeDownloadLocalMenu();
window.testGsi = async () => await testGsi();
window.openFolder = async (a) => await OpenFolder(a);
window.setGsiInSetup = async () => await setGsiInSetup();
window.setupAsgPage = async (a) => await setupAsgPage(a);
window.playVideo = async (d) => await playVideo(d);
window.OpenURL = async (d, b, s) => await OpenURL(d, b, s);
window.addPlugin = async () => await addPlugin();

//PT
window.pton = async function (l) {
    show("main");
    var ptlink = await Pton(l);
    console.log(ptlink);
    u_inp.value = ptlink;
}

window.startAnalyze = async function () {
    clearMain()
    var url;
    a_btn.disabled = true;
    a_inp.disabled = true;
    if (alPlSl != 0) {
        await RunPluginOU(alPlSln, {
            'type': 2,
            'data':  u_inp.value
        });
        a_btn.disabled = false;
        a_inp.disabled = false;
        return
    } else {
        url = u_inp.value;
    }
    var result = "";
    if (url == "") {
        return
    }
    try {
        result = await Analyze(url);
    } catch (err) {
        console.error(err);
        mainError(err);
        a_btn.disabled = false;
        a_inp.disabled = false;
        return 
    }
    try {
        const obj = JSON.parse(result);
        var qii = 0;
        qualitySelectDiv.style.display = "block";

        for (const quality of obj.qualities) {
            qii++
            qualitySelectDiv.insertAdjacentHTML("beforeend", `
            <div class="qualityItem" id="qi${qii}">
                <div class="left">
                    <a class="qualityName">${quality.resolution}</a>
                    <a class="qualityBitrate">${quality.bandwidth}</a>
                </div>
                <div class="right">
                    <button class="button qualitySelectBtn" id="qi${qii}btn" onclick="qualitySelect('${quality.url}', ${qii})">
                        Select
                    </button>
                </div>
            </div>
            `);
        }

        a_btn.disabled = false;
        a_inp.disabled = false;
        a_inp.value = "";
    } catch (error) {
        a_btn.disabled = false;
        a_inp.disabled = false;
        window.qualitySelect(result, 0);
    }
}

window.qualitySelect = function (url, i) {
    if (i != 0) {
        document.querySelectorAll(".qualityItem").forEach(item => {
            if (item.id !== "qi" + i) {
                item.remove();
            }
        });
        document.getElementById("qi" + i + "btn").disabled = true;
        document.getElementById("qi" + i + "btn").innerText = "Selected";
    }

    modeSelectDiv.style.display = "block";

    modeSelectDiv.insertAdjacentHTML("beforeend", `
        <div class="title">Select Mode</div>
        <li><strong>Download</strong> – Download a regular M3U8 video file (VOD).</li>
        <li><strong>Stream</strong> – Watch a live M3U8 stream.</li>
        <li><strong>Watch</strong> – Play the video online without downloading it.</li>
        <div class="modes">
            <button class="button mode" id="ms_download" onclick="modeSelect('${url}', ${i}, 'download')">Download</button>
            <button class="button mode" id="ms_watch" onclick="modeSelect('${url}', ${i}, 'watch')">Watch</button>
            <button class="button mode" id="ms_stream" onclick="modeSelect('${url}', ${i}, 'stream')">Stream</button>
        </div>
    `);
}

window.modeSelect = async function (url, i, mode) {
    console.log("ms_" + mode)
    modeSelectDiv.innerHTML = `<div class="modes" id="modesinmodeselect">
        <button class="button mode" id="ms_download">Download</button>
        <button class="button mode" id="ms_watch">Watch</button>
        <button class="button mode" id="ms_stream">Stream</button>
    </div>`;
    document.getElementById("modesinmodeselect").style.marginTop = 0;
    document.getElementById("ms_" + mode).classList.add("modeSelected");
    PMode = mode;

    abtn(false)
    a_inp.disabled = true;
    a_inp.value = "";

    try {
        const result = await Start(url, mode);
        console.log(result);
        if (result.dtl == "WC") {
            if (mode == "download") {
                playVideo("exports/" + result.dlib.filename);
                var bsed = await bsencoder("exports/" + result.dlib.filename);
                openDownloadLocalMenu(url, bsed);
            }
            abtn(true)
            clearMain();
            a_inp.disabled = false;
        }
    } catch (error) {
        console.log(error);
        mainError(error);
    }
}

window.stop = async (url, i, mode) => await sstop (url, i, mode);

async function sstop (url, i, mode) {
    watchVideo.pause();
    watchVideo.src = "";
    serverRun = false;
    a_btn.disabled = true;
    a_btn.innerText = "Stopping...";
    if (PMode == "watch") {
        await DeleteTs()
        clearMain();
        abtn(true);
        return
    }
    await StopAll();
}

window.runPluginOU = async (a, b) => {
    await RunPluginOU(a, b);
};

window.laChange = function (f) {
    N_la = f;
    laChangeStyle(f);
}

async function getPort() {
    port = await GetPort();
}

function fuinp(data) {
    u_inp.value = data;
    if (alPlSl != 0) {
        alPlSl = 0;
        alPlSln = "normal";
        window.startAnalyze();
    } else {
        show("main");
    }
}

function mainError(err) {
    a_btn.disabled = false;
    a_inp.disabled = false;
    a_inp.value = "";
    let error = String(err);
    if (error.includes("RuntimeError: ")){
        error = error.split("RuntimeError: ")[1];
    }

    switch (error) {
        case "gm:1":
            showError("Connection error");
            break;
        case "gm:2":
            showError("bodyBytes error");
            break;
        case "gm:3":
            showError("AppScript error");
            break;
        case "gm:4":
            showError("Request error");
            break;
        case "gm:5":
            showError("Its not M3U8 file!");
            break;
        case "gm:6":
            showError("Write file error");
            break;

        case "gq:1":
            showError("M3U8 not found!");
            break;
        case "gq:2":
            showError("Open M3U8 file error");
            break;
        case "gq:3":
            showError("Read file error");
            break;
        case "gq:4":
            showError("Quality json marshal error");
            break;

        case "gtn:1":
            showError("M3U8 not found!");
            break;
        case "gtn:2":
            showError("M3U8 Read Error");
            break;
        case "gtn:3":
            showError("Batch Create Error");
            break;

        case "tmie:1":
            showError("Target M3U8 is empty!");
            break;
        case "tmie:2":
            showError("Write final file error");
            break;

        case "port:1":
            showError("Port listen error");
            break;
        default:
            console.log(error);
            break;
    }
}

function showError(error) {
    console.log("errrrrrrrrrrrrrrrrrorrrrrrrrrrrrrrrrrrrr")
    errorbox.innerText = error;
    errorbox.style.display = "block";
    setTimeout(() => {
        errorbox.style.opacity = 1;
    }, 1);
    setTimeout(() => {
        errorbox.style.opacity = 0;
        setTimeout(() => {
            errorbox.style.display = "none";
        }, 250);
    }, 5000);
}

function clearMain() {
    a_btn.disabled = false;
    a_inp.disabled = false;

    progressCount = 0;
    progress = 0;
    progressPer = 0;
    qualitySelectDiv.innerHTML = "";
    modeSelectDiv.innerHTML = "";
    statusDiv.innerHTML = "";
    qualitySelectDiv.style.display = "none";
    modeSelectDiv.style.display = "none";
    statusDiv.style.display = "none";
}

function progressSetup() {
    if (progressCount == 0) {
        serverRun = true;
        statusDiv.style.display = "block";
        statusDiv.innerHTML = `<div class="console">
            <div class="icon">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-chevron-right-icon lucide-chevron-right"><path d="m9 18 6-6-6-6"/></svg>
            </div>
            <div class="cost">Go to the Watch section</div>
        </div>
        <button class="button sabtn" id="a_btn" onclick="show('watch')">
            Watch
        </button>`;
        watch()
    } else {
        statusDiv.style.display = "block";
        statusDiv.innerHTML = `<div class="console">
            <div class="icon">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-chevron-right-icon lucide-chevron-right"><path d="m9 18 6-6-6-6"/></svg>
            </div>
            <div class="cost">Wait for download</div>
        </div>
        <div class="progressDiv">
            <div class="progress" id="progress"></div>
            <div class="data" id="progressData">
                0/${progressCount}
            </div>
        </div>`;
        progressDiv = document.getElementById("progress");
        progressData = document.getElementById("progressData");
        progressDiv.style.width = "0%";
    }
}

function progressUpdate() {
    progressPer = Math.floor((progress * 100) / progressCount);

    progressDiv.style.width = progressPer + "%";
    progressData.innerText = progressPer + "% - " + progress + "/" + progressCount;
}

function abtn(type) {
    if (type == true) {
        a_btn.setAttribute("onclick", "startAnalyze()");
        a_btn.classList.remove("stopbtn");
        a_btn.innerText = "Analyze";
    } else {
        a_btn.setAttribute("onclick", "stop()");
        a_btn.classList.add("stopbtn");
        a_btn.innerText = "Stop";
    }
}

function hideSMExt() {
    smaext.style.display = "none";
}

function hideAll() {
    Cmain.style.display = "none";
    Cwatch.style.display = "none";
    Csetting.style.display = "none";
    Cext.style.display = "none";
    Clib.style.display = "none";

    navBack.classList.remove("navbackUp");
}
function show(c) {
    watchNav(true);
    hideAll();
    document.getElementById("ni_" + Cnow).classList.remove("navActive");
    document.getElementById("ni_" + c).classList.add("navActive");

    Cnow = c
    switch (c) {
        case "main":
            Cmain.style.display = "block";
            slplselect(alPlSl, alPlSln);
            break;
        case "watch":
            Cwatch.style.display = "block";
            watchNavCrt();
            break;
        case "setting":
            Csetting.style.display = "block";
            showNSetting();
            break;
        case "ext":
            Cext.style.display = "block";
            getPluginList();
            break;
        case "lib":
            Clib.style.display = "block";
            libLoad();
            break;
        default:
            Cmain.style.display = "block";
            break;
    }
}
function watchNavCrt() {
    navBack.classList.add("navbackUp");
    watchNav(true);
    watchBtnNav(false);
    setTimeout(() => {
        if (Cnow != "watch") return;
        watchNav(false);
        watchBtnNav(true);
    }, 5000);
}
function watchNav(t) {
    
    if (t) {
        navBack.style.display = "flex";
        setTimeout(() => {
            navBack.style.transform = "translateY(0px)"
        }, 10);
    } else {
        navBack.style.transform = "translateY(-100px)";
        setTimeout(() => {
            navBack.style.display = "none";
        }, 250);
    }
}
function watchBtnNav(t) {
    
    if (t) {
        watchNavBtnB.style.display = "flex";
        setTimeout(() => {
            watchNavBtnB.style.transform = "translateY(0px)"
        }, 10);
    } else {
        watchNavBtnB.style.transform = "translateY(-40px)";
        setTimeout(() => {
            watchNavBtnB.style.display = "none";
        }, 250);
    }
}

function watch() {
    getPort()
    serverRunCheck()

    if (Hls.isSupported()) {
        const hls = new Hls({
            enableWorker: true,
            lowLatencyMode: true,
        });

        hls.loadSource("/watch");
        hls.attachMedia(watchVideo);

        hls.on(Hls.Events.MANIFEST_PARSED, () => {
            watchVideo.play();
        });
    } else if (video.canPlayType("application/vnd.apple.mpegurl")) {
        watchVideo.src = "/watch";
        watchVideo.play();
        
    }
}
async function playVideo(tsrc) {
    console.log("sssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssss");
    if (serverRun) {
        await sstop(null, null, null);
        a_btn.disabled = false;
        clearMain();
        abtn(true);
    }
    serverRun = true;
    const bsed = await bsencoder(tsrc);

    show("watch");

    watchVideo.oncanplay = async () => {
        watchVideo.oncanplay = null;
        await watchVideo.play();
    };
    elmvideo.src = "/pret/" + bsed;
    console.log("/pret/" + bsed);
    elmvideo.load();
    // watchVideo.addEventListener("error", () => console.log(watchVideo.error));
    
}

function serverRunCheck() {
    if (serverRun) {
        notRun.style.display = "none";
    } else {
        watchVideo.pause();
        watchVideo.src = "";
        notRun.style.display = "flex";
    }
}


async function getConfig() {
    const cfg = await GetConfig();
    gsa = cfg.appScriptKey;
    bs = cfg.batchSize;
    ms = cfg.maxSize;
    la = cfg.lowLatency;
    N_gsa = gsa;
    N_bs = bs;
    N_ms = ms;
    N_la = la;
}
async function updateConfig() {
    await UpdateConfig(N_gsa, parseInt(N_bs), parseInt(N_ms), N_la);
    await getConfig();
    showNSetting();
}
function showNSetting(){
    s_gsa.value = N_gsa;
    s_bs.value = N_bs;
    s_ms.value = N_ms;
    laChangeStyle(N_la);
    checkSettingChange()
}
function laChangeStyle(f) {
    if (f) {
        s_la_on.classList.add("chon");
        s_la_off.classList.remove("chon");
    } else {
        s_la_off.classList.add("chon");
        s_la_on.classList.remove("chon");
    }
}
function nChange() {
    if (!s_gsa.reportValidity() || !s_bs.reportValidity() || !s_ms.reportValidity()) {
        return;
    }
    if (
        !s_gsa.reportValidity() ||
        !s_bs.reportValidity() ||
        !s_ms.reportValidity() ||
        s_bs.value === "" ||
        s_ms.value === "" ||
        isNaN(s_bs.value) ||
        isNaN(s_ms.value) ||
        s_bs.value <= 0 ||
        s_ms.value <= 0
    ) {
        return;
    }

    N_gsa = s_gsa.value;
    N_bs = s_bs.value;
    N_ms = s_ms.value;
    
    checkSettingChange()
}
function checkSettingChange() {
    if (N_gsa != gsa || N_bs != bs || N_ms != ms || N_la != la) {
        saveChangesBtn.disabled = false
    } else {
        saveChangesBtn.disabled = true
    }
}

async function getPluginList() {
    let list = await GetPluginListOU();
    if (list == null) {
        console.log("gpl NULL");
        return;
    }

    if (list.length <= 0) {
        addPluginBtn.innerText = "Add plugin";
        MaddPluginBtn.innerText = "Add plugin";
    }

    deleteAllExtButton();
    for (let i = 0; i < list.length; i++) {
        var ii = i + 1;
        let pluginName = list[i].split(".js")[0];
        let onclickPluginStart = `slplsel(${ii}, '${pluginName}')`;
        let onclickMPluginStart = `runPluginOU('` + pluginName + `', {
        'type': 1,
        'data': ''
        }); selectPluginStyle('plbtn` + ii + `'); hideSMExt();`;
        MextList.insertAdjacentHTML("beforeend", `<div class="mextItem" onclick="` + onclickPluginStart + `" id="Mplbtn` + ii + `">` + pluginName + `</div>`);
        extList.insertAdjacentHTML("beforeend", `<div class="extItem" onclick="` + onclickMPluginStart + `" id="plbtn` + ii + `">` + pluginName + `</div>`);
    }
}
function selectPluginStyle(id) {
    unselectAllPluginStyle()
    console.log(id + "sssssssssssssssssssssssss")
    document.getElementById(id).style.background = "var(--surface-3)";
}
function unselectAllPluginStyle() {
    console.log("unsssssssssssse")
    extList.childNodes.forEach(node => {
        node.style.background = "var(--bg)";
    });
    MextList.childNodes.forEach(node => {
        node.style.background = "var(--bg)";
    });
}
function deleteAllExtButton() {
    Array.from(extList.childNodes).forEach(node => {
        if (node.id != "addPluginBtn" && node.id != "Mplbtn0") {
            console.log(node.id);
            node.remove();
        }
    });
    Array.from(MextList.childNodes).forEach(node => {
        if (node.id != "addPluginBtn_m" && node.id != "Mplbtn0") {
            console.log(node.id);
            node.remove();
        }
    });
}

async function addPlugin() {
    const newPath = await AddPluginOU();
    console.log(newPath);    
    getPluginList()
}

function slplselect(i, name) {
    getPluginList()
    setTimeout(() => {
        selectPluginStyle("Mplbtn" + i);
    }, 100);
    
    console.log(i + name)
    alPlSl = i;
    alPlSln = name;
}

async function libLoad() {
    const libs = await LibLoad();
    Clib.innerHTML = "";
    for (const lib of libs) {
        if (lib.type == "download") {
            var bsed = await bsencoder("exports/" + lib.filename);
            Clib.insertAdjacentHTML("beforeend", `<div class="litem">
                <div class="buttons">
                    <a class="lbtn" onclick="fuinp('${lib.url}')">
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="96" height="96">
                            <path fill-rule="evenodd" d="M20 12C20 16.417969 16.417969 20 12 20C7.582031 20 4 16.417969 4 12C4 7.582031 7.582031 4 12 4C13.113281 4 14.167969 4.238281 15.132813 4.648438L12.628906 7.324219L20.121094 7.074219L19.484375 0L17.273438 2.359375C15.710938 1.5 13.914063 1 12 1C5.925781 1 1 5.925781 1 12C1 18.074219 5.925781 23 12 23C18.074219 23 23 18.074219 23 12Z"/>
                        </svg>
                    </a>
                    <a class="lbtn">
                        <svg xmlns="http://www.w3.org/2000/svg" onclick="openDownloadLocalMenu('${lib.url}', '${bsed}')" viewBox="0 0 24 24" width="96" height="96">
                            <path d="M11 0L11 19.5625L2.71875 11.28125L1.28125 12.71875L11.28125 22.71875L12 23.40625L12.71875 22.71875L22.71875 12.71875L21.28125 11.28125L13 19.5625L13 0Z"/>
                        </svg>
                    </a>
                    <a class="lbtn" onclick="playVideo('exports/${lib.filename}')">
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="96" height="96">
                            <path d="M4 2L4 22L21.3125 12Z" />
                        </svg>
                    </a>
                </div>
                <div class="info">
                    <a class="url">${truncate(lib.url, 15)}</a>
                    <a class="time">${formatDate(lib.date)} | ${lib.type}</a>
                </div>
                <div class="poster"></div>
            </div>`);
        } else {
            Clib.insertAdjacentHTML("beforeend", `<div class="litem">
                <div class="buttons">
                    <a class="lbtn" onclick="fuinp('${lib.url}')">
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="96" height="96">
                            <path fill-rule="evenodd" d="M20 12C20 16.417969 16.417969 20 12 20C7.582031 20 4 16.417969 4 12C4 7.582031 7.582031 4 12 4C13.113281 4 14.167969 4.238281 15.132813 4.648438L12.628906 7.324219L20.121094 7.074219L19.484375 0L17.273438 2.359375C15.710938 1.5 13.914063 1 12 1C5.925781 1 1 5.925781 1 12C1 18.074219 5.925781 23 12 23C18.074219 23 23 18.074219 23 12Z"/>
                        </svg>
                    </a>
                </div>
                <div class="info">
                    <a class="url">${truncate(lib.url, 15)}</a>
                    <a class="time">${formatDate(lib.date)} | ${lib.type}</a>
                </div>
                <div class="poster"></div>
            </div>`);
        }
    }
    Clib.insertAdjacentHTML("beforeend", `<br><br><br><br>`);
}

function truncate(text, maxLength) {
    if (text.length <= maxLength) return text;
    return text.slice(0, Math.max(0, maxLength - 3)) + "...";
}
function formatDate(isoDate) {
    const d = new Date(isoDate);

    const pad = n => String(n).padStart(2, "0");

    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
async function bsencoder(v) {
    return await BSEncoder(v);
}


async function setupAsgPage(id) {
    for (let i = 0; i <= 3; i++) {
        if (i == id) continue;
        let hp = document.getElementById("setupasg" + i);
        hp.style.opacity = 0;
        await sleep(250);
        hp.style.display = "none";
    }

    var p = document.getElementById("setupasg" + id);

    p.style.display = "block";
    setTimeout(() => {
        p.style.opacity = 1;
    }, 10);
}

function copy(type) {
    switch (type) {
        case 1:
            Copy(`/*
===========================================================
Totem - Google Apps Script Relay Setup
===========================================================

1. Open https://script.google.com
2. Create a new Apps Script project.
3. Delete the default code and paste this entire file.
4. Click "Deploy" → "New deployment".
5. Select "Web app".
6. Configure:
   - Execute as: Me
   - Who has access: Anyone
7. Click "Deploy" and authorize the requested permissions.
8. Copy the generated Web App KEY.
9. Open App and put this in setting.

===========================================================
*/

function doGet(e) {
  const type = e.parameter.type
  var headers = {};
  if (e.parameter.h) {
    headers = JSON.parse(e.parameter.h);
  }
  if(type == 0) {
    try {
      var url = e.parameter.url;
      var response = UrlFetchApp.fetch(url, {
        'method': 'get',
        'muteHttpExceptions': true,
        'headers': headers
      });

      var statusCode = response.getResponseCode();

      if (statusCode == 200) {
        return ContentService.createTextOutput(response.getContentText())
          .setMimeType(ContentService.MimeType.TEXT);
      } else {
        return ContentService.createTextOutput('Error: Status Code ' + statusCode)
          .setMimeType(ContentService.MimeType.TEXT);
      }
      
    } catch (e) {
      return ContentService.createTextOutput('Error: ' + e.toString())
        .setMimeType(ContentService.MimeType.TEXT);
    }
  } else if(type == 1) {
    try {
      var url = e.parameter.url;
      var response = UrlFetchApp.fetch(url, {
        'method': 'get',
        'muteHttpExceptions': true,
        'headers': headers
      });

      var statusCode = response.getResponseCode();

      var blob = response.getBlob();
      var bytes = blob.getBytes();

      var hex = bytes
        .map(function(b) {
          return ('0' + ((b + 256) % 256).toString(16)).slice(-2);
        })
        .join(' ');
      
      if (statusCode == 200) {
        return ContentService.createTextOutput(hex)
          .setMimeType(ContentService.MimeType.TEXT);
      } else {
        return ContentService.createTextOutput('Error: Status Code ' + statusCode)
          .setMimeType(ContentService.MimeType.TEXT);
      }
      
    } catch (e) {
      return ContentService.createTextOutput('Error: ' + e.toString())
        .setMimeType(ContentService.MimeType.TEXT);
    }
  }
}

function doPost(e){
  const data = JSON.parse(e.postData.contents);
  
  try {
    const response = UrlFetchApp.fetch(
      data.url,
      {
        method: "post",
        contentType: "application/json",
        headers: data.headers || {},
        payload: JSON.stringify(data.payload),
        muteHttpExceptions: true
      }
    );

    var statusCode = response.getResponseCode();

    if (statusCode == 200) {
      return ContentService.createTextOutput(response.getContentText())
        .setMimeType(ContentService.MimeType.TEXT);
    } else {
      return ContentService.createTextOutput('Error: Status Code ' + statusCode)
        .setMimeType(ContentService.MimeType.TEXT);
    }
  } catch (e) {
    return ContentService.createTextOutput('Error: ' + e.toString())
      .setMimeType(ContentService.MimeType.TEXT);
  }
}
`);
            break;

        case 2:
            Copy("https://script.google.com");
            break;
        default:
            break;
    }
}

async function testGsi() {
    var v = testGsiInp.value;
    if (v == "") return;

    gsiStatusTxt.style.color = "var(--text)";
    gsiStatusTxt.innerText = "Connecting...";

    var result = await TestGsi(v);
    var resultTxt = "";
    if (result == "OK") {
        gsiStatusTxt.style.color = "var(--success)";
        resultTxt = "Connected!"
    } else {
        gsiStatusTxt.style.color = "var(--error)";

        switch (result) {
            case "cgsi:1":
                resultTxt = "Could not connect to Google.";
                break;

            case "cgsi:2":
                resultTxt = "The connection to the specified app script key could not be established.";
                break;
        }
    }

    gsiStatusTxt.innerText = resultTxt;
}
async function setGsiInSetup(){
    var v = testGsiInp.value;
    if (v == "") return;
    await UpdateConfig(v, parseInt(bs), parseInt(ms), la);
    await getConfig();
    document.getElementById("setupasg" + 0)
    document.getElementById("setupasg" + 0).style.display = "block";
    setupback.style.display = "none";
}

function openDownloadLocalMenu(url, tsed) {
    let rurl = "localhost:1820/tsb/" + tsed;
    if (!isMobile) {
        m_dwpb_filename.innerText = truncate(url, 15);
        m_dwpb_code.innerText = rurl;
        m_dwpb_copy.setAttribute("onclick", "copy('" + rurl + "')");
        m_dwpb_opbtn.setAttribute("onclick", "openFolder('dpex')");
        dwpathboxm.style.display = "flex";
    } else {
        w_dwpb_filename.innerText = truncate(url, 15);
        w_dwpb_code.innerText = rurl;
        w_dwpb_copy.setAttribute("onclick", "copy('" + rurl + "')");
        dwpathboxw.style.display = "flex";
    }
}
function closeDownloadLocalMenu() {
    dwpathboxm.style.display = "none";
    dwpathboxw.style.display = "none";
}