const slides = [...document.querySelectorAll(".slide")];
    const pos = document.getElementById("pos") || {};
    const gate = document.getElementById("gate");
    const params = new URLSearchParams(location.search);
    const auto = params.has("record");
    let i = Math.max(0, Math.min(slides.length - 1, Number(params.get("s") || 0)));
    let started = false;
    let asking = false;
    let askVo = false;
    let resuming = false;
    let askTimer = 0;
    let askLeft = 30;
    const askNames = {
      single: "Paused · live",
      multi: "Paused · live",
      quiz: "Paused · live",
      code: "Paused · live",
    };
    const pauseClips = ["vo/pause-1.mp3", "vo/pause-2.mp3"];
    const resumeClips = ["vo/resume-1.mp3", "vo/resume-2.mp3", "vo/resume-3.mp3"];
    let liveBeat = 0;
    const ask = document.getElementById("ask");
    const askClock = document.getElementById("ask-clock");

    const vo = new Audio();
    const sfx = {
      whoosh: new Audio("vo/sfx-whoosh.mp3"),
      hit: new Audio("vo/sfx-hit.mp3"),
      stamp: new Audio("vo/sfx-stamp.mp3"),
      tick: new Audio("vo/sfx-tick.mp3"),
      rise: new Audio("vo/sfx-rise.mp3"),
    };
    Object.values(sfx).forEach((a) => { a.preload = "auto"; a.volume = 0.34; });
    vo.preload = "auto";
    vo.preservesPitch = true;
    vo.playbackRate = 0.9;

    const grid = document.getElementById("grid-a");
    const cells = [];
    if (grid) {
      for (let n = 0; n < 18 * 12; n++) {
        const c = document.createElement("div");
        c.className = "cell";
        grid.appendChild(c);
        cells.push(c);
      }
    }

    function playSfx(name) {
      const a = sfx[name];
      if (!a) return;
      a.pause();
      a.currentTime = 0;
      a.play().catch(() => {});
    }

    function lightGrid() {
      cells.forEach((c) => c.classList.remove("lit"));
      const lit = 18 * 4;
      let n = 0;
      const step = () => {
        if (n >= lit) return;
        cells[n].classList.add("lit");
        if (n % 8 === 0) playSfx("tick");
        n += 1;
        setTimeout(step, 22);
      };
      step();
    }

    function restartAnims(el) {
      el.querySelectorAll(".in, .stamp-in, .badge-in, .fill, .col, .rule, .ill-coaster").forEach((node) => {
        node.style.animation = "none";
        void node.offsetWidth;
        node.style.animation = "";
      });
    }

    function show(n) {
      const nextI = Math.max(0, Math.min(slides.length - 1, n));
      slides.forEach((s, idx) => {
        s.classList.toggle("exit", idx === i && idx !== nextI);
        s.classList.toggle("on", idx === nextI);
        if (idx !== nextI && idx !== i) s.classList.remove("exit");
      });
      i = nextI;
      if ("textContent" in pos) pos.textContent = `${i + 1} / ${slides.length}`;
      const slide = slides[i];
      restartAnims(slide);
      slide.querySelector("animateMotion")?.beginElement?.();
      playSfx(slide.dataset.enter || "whoosh");
      if (slide.dataset.grid) setTimeout(lightGrid, 180);
      if (slide.dataset.stamp) setTimeout(() => playSfx("stamp"), Number(slide.dataset.stamp));
      if (slide.dataset.rise) setTimeout(() => playSfx("rise"), Number(slide.dataset.rise));
      vo.pause();
      if (slide.dataset.vo) {
        askVo = false;
        vo.src = slide.dataset.vo;
        vo.currentTime = 0;
        vo.playbackRate = 0.9;
        vo.play().catch(() => {});
      }
    }

    function next() {
      if (asking || resuming) return;
      if (i < slides.length - 1) show(i + 1);
    }

    function openAsk() {
      const kind = slides[i].dataset.ask;
      if (!kind || !ask) return;
      asking = true;
      resuming = false;
      ask.classList.remove("out");
      ask.querySelectorAll(".ask-pane").forEach((p) => p.classList.toggle("on", p.dataset.kind === kind));
      document.getElementById("ask-kind").textContent = askNames[kind] || "Paused · live";
      askLeft = 30;
      askClock.textContent = askLeft;
      ask.classList.add("on");
      playSfx("stamp");
      askVo = true;
      vo.pause();
      vo.src = pauseClips[liveBeat % pauseClips.length];
      vo.currentTime = 0;
      vo.playbackRate = 0.9;
      vo.play().catch(() => {});
      clearInterval(askTimer);
      askTimer = setInterval(() => {
        askLeft -= 1;
        askClock.textContent = Math.max(0, askLeft);
        if (askLeft <= 0) closeAsk(true);
      }, 1000);
    }

    function closeAsk(advance) {
      if (!asking || resuming) return;
      asking = false;
      clearInterval(askTimer);
      if (!advance) {
        ask.classList.remove("on", "out");
        return;
      }
      resuming = true;
      askVo = false;
      ask.classList.add("out");
      playSfx("whoosh");
      vo.pause();
      vo.src = resumeClips[liveBeat % resumeClips.length];
      liveBeat += 1;
      vo.currentTime = 0;
      vo.playbackRate = 0.9;
      vo.play().catch(() => {
        resuming = false;
        ask.classList.remove("on", "out");
        next();
      });
    }

    vo.addEventListener("ended", () => {
      if (askVo) {
        askVo = false;
        return;
      }
      if (resuming) {
        resuming = false;
        ask.classList.remove("on", "out");
        next();
        return;
      }
      if (started && slides[i].dataset.ask) openAsk();
      else if (started || auto) setTimeout(next, 220);
    });
    (ask ? ask.querySelectorAll("[data-kind=single] button") : []).forEach((btn) => {
      btn.addEventListener("click", () => {
        (ask ? ask.querySelectorAll("[data-kind=single] button") : []).forEach((b) => {
          b.classList.toggle("pick", b === btn);
          b.classList.toggle("ok", b === btn && b.dataset.ok);
          b.classList.toggle("no", b === btn && !b.dataset.ok);
        });
      });
    });
    (ask ? ask.querySelectorAll("[data-kind=multi] label") : []).forEach((lab) => {
      lab.addEventListener("click", () => lab.classList.toggle("pick"));
    });
    (ask ? ask.querySelectorAll("[data-quiz]") : []).forEach((btn) => {
      btn.addEventListener("click", () => {
        (ask ? ask.querySelectorAll("[data-quiz]") : []).forEach((b) => {
          b.classList.toggle("ok", b.dataset.quiz === "ok");
          b.classList.toggle("no", b.dataset.quiz === "no");
          b.classList.toggle("pick", b === btn);
        });
      });
    });
    document.getElementById("ghost")?.addEventListener("click", () => {
      document.getElementById("ghost").classList.add("in");
    });
    document.getElementById("ask-go")?.addEventListener("click", () => closeAsk(true));

    function start() {
      if (started) return;
      started = true;
      gate?.classList.add("hide");
      show(i);
    }

    gate?.addEventListener("click", start);
    if (auto) {
      gate?.classList.add("hide");
      document.querySelector(".hud")?.classList.add("hide");
      setTimeout(start, Number(params.get("wait") || 1600));
    }
    document.addEventListener("keydown", (e) => {
      if (e.key === "Enter" || e.key === " ") {
        if (!started) { e.preventDefault(); start(); return; }
      }
      if (!started) return;
      if (asking && e.key === "Tab") {
        e.preventDefault();
        document.getElementById("ghost")?.classList.add("in");
        return;
      }
      if (asking && (e.key === "Enter" || e.key === "Escape")) {
        e.preventDefault();
        closeAsk(true);
        return;
      }
      if (asking) return;
      if (["ArrowRight", "PageDown"].includes(e.key)) { e.preventDefault(); next(); }
      if (["ArrowLeft", "PageUp"].includes(e.key)) { e.preventDefault(); show(i - 1); }
      if (e.key === "f" || e.key === "F") {
        if (!document.fullscreenElement) document.documentElement.requestFullscreen();
        else document.exitFullscreen();
      }
    });
