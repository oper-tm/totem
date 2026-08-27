const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));

msl();
async function l1Ch() {
    let l1c = document.getElementById("l1");
    await changeSlide("l11", "l13", true, l1c, "110px");
    await sleep(2000);
    await changeSlide("l12", "l11", true, l1c, "72px");
    await sleep(2000);
    await changeSlide("l13", "l12", true, l1c, "82px");
    await sleep(2000);
}
async function msl() {
    while (true) {
        changeSlide("us1", "us4");
        await l1Ch();
        await changeSlide("us2", "us1");
        await sleep(5000);
        await changeSlide("us3", "us2");
        await sleep(5000);
        await changeSlide("us4", "us3");
        await sleep(5000);
    }
}
async function changeSlide(now, past = null, tr = false, c = document.getElementById(""), w = "none") {
    if (past != null) {
        let d = document.getElementById(past);
        d.style.opacity = 0;
        d.style.transform = "translateY(100%)";
        await sleep(500);
        d.style.display = "none"
        if (tr) d.style.transform = "translateY(-100%)";
    }
    if (now != null) {
        let d = document.getElementById(now);
        if (tr) d.style.transform = "translateY(-100%)";
        d.style.display = "flex"
        await sleep(500);
        d.style.opacity = 1;
        d.style.transform = "translateY(0)";
        if (c != null) c.style.width = w;
    }
}
async function changel1Slide(now, past = null, tr = true) {
    if (past != null) {
        let d = document.getElementById(past);
        d.style.opacity = 0;
        d.style.transform = "translateY(100%)";
        await sleep(500);
        d.style.display = "none"
    }
    if (now != null) {
        let d = document.getElementById(now);
        d.style.display = "flex"
        await sleep(500);
        d.style.opacity = 1;
        d.style.transform = "translateY(0)";
        
    }
}