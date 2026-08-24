host.log("Xnxx started");

switch (args.type) {
    case 1:
        getMainList();
        break;
    case 2:
        getVideoList(args.data)
        break;
    case 3:
        getVideo(args.data)
        break;
    default:
        break;
}

function getMainList() {
    host.loadingPage(true);
    let mainPage = host.getUrl("https://xnxx.com");
    const stringListJson = mainPage.split('cats.write_thumb_block_list(')[1].split(', "home-cat-list");')[0];

    let items = [];
    let videoListLink = '';

    var fenum = 0;

    try {
        items = JSON.parse(stringListJson);
    } catch (err) {
        host.log("Categories unmarshal error:", err);
        return [err, "err"];
    }

    items.forEach((item, i) => {
        fenum++;
        let onclickfun = `runPluginOU('test', {
            'type': 2,
            'data': '` + item.u + `'
        })`
        if (fenum == 1) {
            host.II_A(item.tf, "decv", onclickfun, true);
        } else {
            host.II_A(item.tf, "devc", onclickfun, false);
        }
        host.log(item.tf);
    });
    host.loadingPage(false);
}

function getVideoList(vu) {
    host.loadingPage(true);
    let mainPage = host.getUrl("https://xnxx.com" + vu);
    let vls = mainPage.replace(/\r/g, "").replace(/\n/g, "");
    let exp1 = vls.split('<div id="video_');
    var fenum = 0;
    for (let i = 0; i < exp1.length; i++) {
        if (i === 0) continue;
        fenum++;
        let video = exp1[i].split('</span></p></div></div>')[0];
        let Xs_videoID = video.split('"')[0];
        let Xs_videoDataID = video.split('data-id="')[1].split('"')[0];
        let Xs_videoDataEID = video.split('data-eid="')[1].split('"')[0];
        let Xs_videoHref = "https://xnxx.com" + video.split('<div class="thumb"><a href="')[1].split('"')[0];
        let Xs_videoThumb = video.split('data-src="')[1].split('"')[0];
        let Xs_videoTitle = video.split('title="')[1].split('"')[0];
        let Xs_videoPub = "None";
        if (video.includes('<span class="name">')) {
            Xs_videoPub = video.split('<span class="name">')[1].split('</span>')[0];
        }
        let Xs_videoView = video.split('<span class="right">')[1].split(' <span ')[0];
        let Xs_videoDuration = video.split('</span></span>')[1].split('<span ')[0];
        
        let onclickfun = `runPluginOU('test', {
            'type': 3,
            'data': '` + Xs_videoHref + `'
        })`
        if (fenum == 1) {
            host.II_B(host.base64(Xs_videoThumb), Xs_videoTitle, Xs_videoPub, onclickfun, true);
        } else {
            host.II_B(host.base64(Xs_videoThumb), Xs_videoTitle, Xs_videoPub, onclickfun, false);
        }
        host.log("[" + i + "]" + Xs_videoDuration + ">" + Xs_videoView + ">" + Xs_videoPub + ">" + Xs_videoTitle);
    }
    host.loadingPage(false)
}

function getVideo(vu) {
    host.loadingPage(true);
    let mainPage = host.getUrl(vu);
    let videoM3u8 = mainPage.split("html5player.setVideoHLS('")[1].split("');")[0];
    host.loadingPage(false);
    host.setInp(videoM3u8);
}