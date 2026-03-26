package server

import (
	"html/template"
	"strconv"
)

var scoreLabels = []string{
	"ESCAPED CLASSIFICATION",
	"LEVEL 3 WOOK",
	"LEVEL 2 WOOK",
	"LEVEL 1 WOOK",
	"NORMIE",
	"LEVEL 1 WOKE",
	"LEVEL 2 WOKE",
	"LEVEL 3 WOKE",
}

var scoreVibes = []string{
	"Escaped classification entirely",
	"Raw, earthy, chaotic, festival energy",
	"Warm, cloudy, irregular, rough",
	"Leaning wook",
	"Dead center, neither wook nor woke",
	"Leaning woke",
	"Cool, clear, geometric, polished",
	"Polished, geometric, precise, museum-ready",
}

var tmplFuncs = template.FuncMap{
	"scoreLabel": func(s int) string {
		if s < 0 || s > 7 {
			return "OFF THE CHARTS"
		}
		return scoreLabels[s]
	},
	"scoreVibe": func(s int) string {
		if s < 0 || s > 7 {
			return "Score outside the 0–7 range"
		}
		return scoreVibes[s]
	},
	"mul": func(a, b int) int { return a * b },
	// scorePos maps scores 1–7 across 0–100%, score 0 returns -1 (hidden)
	"scorePos": func(s int) int {
		if s < 1 || s > 7 {
			return -1
		}
		return (s - 1) * 100 / 6
	},
	// votePos maps a float avg (1.0–7.0) to 0–100%
	"votePos": func(avg float64) int {
		if avg <= 1 {
			return 0
		}
		if avg >= 7 {
			return 100
		}
		return int(((avg - 1) / 6) * 100)
	},
	"printf1f": func(f float64) string {
		return strconv.FormatFloat(f, 'f', 1, 64)
	},
}

var loginTmpl = template.Must(template.New("login").Parse(`<!DOCTYPE html>
<html><head>
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>wook2woke - Login</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: system-ui; background: #0a0a0a; color: #e0e0e0; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
  .login { background: #1a1a1a; padding: 2rem; border-radius: 12px; width: 90%; max-width: 360px; }
  h1 { font-size: 1.5rem; margin-bottom: 1rem; text-align: center; }
  input { width: 100%; padding: 0.75rem; margin-bottom: 1rem; border: 1px solid #333; border-radius: 8px; background: #111; color: #e0e0e0; font-size: 1rem; }
  button { width: 100%; padding: 0.75rem; border: none; border-radius: 8px; background: #4f46e5; color: white; font-size: 1rem; cursor: pointer; }
  button:hover { background: #4338ca; }
  .error { color: #ef4444; margin-bottom: 1rem; text-align: center; }
</style>
</head><body>
<div class="login">
  <h1>wook2woke</h1>
  {{if .Error}}<p class="error">{{.Error}}</p>{{end}}
  <form method="POST" action="/login">
    <input type="password" name="password" placeholder="Password" autofocus>
    <button type="submit">Enter</button>
  </form>
</div>
</body></html>`))

var liveTmpl = template.Must(template.New("live").Funcs(tmplFuncs).Parse(`<!DOCTYPE html>
<html><head>
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>wook2woke — live</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: system-ui; background: #0a0a0a; color: #e0e0e0; height: 100vh; overflow: hidden; display: flex; flex-direction: column; }
  /* ── Analyzing overlay ── */
  #analyzing { position: fixed; inset: 0; z-index: 50; background: #0a0a0a; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 2rem; transition: opacity 0.4s; }
  #analyzing.hidden { opacity: 0; pointer-events: none; }
  .scan-label { font-size: 0.75rem; letter-spacing: 0.25em; text-transform: uppercase; color: #444; }
  #analysis-msg { font-size: 1.6rem; font-weight: 700; letter-spacing: 0.05em; text-align: center; max-width: 600px; color: #e0e0e0; min-height: 2.5rem; }
  .progress-track { width: 320px; height: 4px; background: #1a1a1a; border-radius: 2px; overflow: hidden; }
  #progress-fill { height: 100%; width: 0%; border-radius: 2px; background: linear-gradient(90deg, #f97316, #818cf8); transition: width 0.15s linear; }
  .blink { animation: blink 1s step-end infinite; }
  @keyframes blink { 50% { opacity: 0; } }
  /* ── Main display ── */
  #display { flex: 1; display: grid; grid-template-columns: 2fr 1fr; height: 100vh; opacity: 0; transition: opacity 0.6s; }
  #display.visible { opacity: 1; }
  #photo-side { position: relative; overflow: hidden; background: #111; display: flex; align-items: center; justify-content: center; }
  #photo-side img { width: 100%; height: 100%; object-fit: cover; display: block; }
  #info-side { display: flex; flex-direction: column; justify-content: center; padding: 4rem 3rem; gap: 1.5rem; background: #0d0d0d; }
  @media (max-width: 640px) {
    #display { grid-template-columns: none; grid-template-rows: 2fr 1fr; }
    #info-side { padding: 1.2rem 1.4rem; gap: 0.75rem; }
    #live-level { font-size: 1.4rem; padding: 0.35rem 0.8rem; }
    #live-desc { font-size: 0.95rem; max-width: 100%; }
    #live-roast-label { font-size: 0.75rem; }
    #live-roast-text { font-size: 0.85rem; }
    .spectrum-wrap { gap: 0.2rem; }
  }
  #live-score { display: none; }
  #live-level { font-size: 2.8rem; font-weight: 900; letter-spacing: 0.06em; text-transform: uppercase; border: 1px solid; border-radius: 10px; padding: 0.5rem 1.2rem; display: inline-block; width: fit-content; }
  #live-desc { font-size: 1.5rem; line-height: 1.5; color: #ccc; max-width: 480px; }
  .spectrum-wrap { display: flex; flex-direction: column; gap: 0.4rem; }
  .spectrum-labels { display: flex; justify-content: space-between; font-size: 0.7rem; letter-spacing: 0.1em; text-transform: uppercase; }
  .spectrum-labels .lbl-wook { color: #f97316; }
  .spectrum-labels .lbl-woke { color: #818cf8; }
  .spectrum-track { position: relative; height: 6px; border-radius: 3px; background: linear-gradient(90deg, #f97316, #fbbf24, #6b7280, #60a5fa, #818cf8); overflow: visible; }
  #live-vote .vote-label { font-size: 0.75rem; font-weight: 700; letter-spacing: 0.15em; text-transform: uppercase; color: #ccc; margin-bottom: 0.3rem; display: flex; justify-content: space-between; align-items: baseline; }
  #live-vote #live-vstats { font-size: 0.75rem; color: #888; font-weight: 400; letter-spacing: 0; text-transform: none; }
  #live-vote .vote-slider { position: absolute; width: 100%; left: 0; top: 50%; transform: translateY(-50%); height: 64px; background: transparent; margin: 0; padding: 0; -webkit-appearance: none; appearance: none; cursor: pointer; z-index: 1; }
  #live-vote .vote-slider::-webkit-slider-runnable-track { background: transparent; height: 6px; }
  #live-vote .vote-slider::-moz-range-track { background: transparent; height: 6px; }
  #live-vote .vote-slider::-webkit-slider-thumb { -webkit-appearance: none; width: 28px; height: 28px; border-radius: 50%; background: white; cursor: pointer; border: 2px solid #0d0d0d; box-shadow: 0 0 8px rgba(255,255,255,0.6); margin-top: -11px; }
  #live-vote .vote-slider::-moz-range-thumb { width: 28px; height: 28px; border-radius: 50%; background: white; cursor: pointer; border: 2px solid #0d0d0d; }
  #spectrum-dot { position: absolute; top: 50%; width: 14px; height: 14px; border-radius: 50%; background: white; border: 2px solid #0d0d0d; transform: translate(-50%, -50%); transition: left 0.8s cubic-bezier(.34,1.56,.64,1); box-shadow: 0 0 8px rgba(255,255,255,0.6); }
  #spectrum-unknown { position: absolute; top: 50%; transform: translateY(-50%) translateX(-50%); display: none; align-items: center; justify-content: center; animation: scan 5s ease-in-out infinite; }
  #spectrum-unknown span { font-size: 4rem; font-weight: 700; color: white; display: block; animation: wobble 4s ease-in-out infinite; }
  @keyframes scan {
    0%   { left: 0%; }
    50%  { left: 100%; }
    100% { left: 0%; }
  }
  @keyframes wobble {
    0%   { transform: rotate(-14deg) scale(1); }
    25%  { transform: rotate(11deg)  scale(2); }
    50%  { transform: rotate(-9deg)  scale(1); }
    75%  { transform: rotate(13deg)  scale(2); }
    100% { transform: rotate(-14deg) scale(1); }
  }
  #back-btn { position: fixed; top: 0.75rem; right: 0.75rem; z-index: 100; font-size: 0.72rem; letter-spacing: 0.1em; text-transform: uppercase; color: #555; text-decoration: none; border: 1px solid #222; border-radius: 6px; padding: 0.25rem 0.6rem; background: rgba(10,10,10,0.7); }
  #back-btn:hover { color: #aaa; border-color: #444; }
  #waiting { position: fixed; inset: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 1rem; }
  #waiting h2 { font-size: 1.4rem; color: #444; letter-spacing: 0.1em; text-transform: uppercase; }
  .s0 { color: #9ca3af; border-color: #4b5563; background: #1c1f26; }
  .s1 { color: #fb923c; border-color: #9a3412; background: #2c1008; }
  .s2 { color: #fbbf24; border-color: #b45309; background: #2c1a04; }
  .s3 { color: #fde68a; border-color: #ca8a04; background: #2a1f03; }
  .s4 { color: #d1d5db; border-color: #6b7280; background: #1f2937; }
  .s5 { color: #7dd3fc; border-color: #0369a1; background: #082030; }
  .s6 { color: #93c5fd; border-color: #1d4ed8; background: #0d1f42; }
  .s7 { color: #c4b5fd; border-color: #6d28d9; background: #1e0a40; }
</style>
</head><body>
<a id="back-btn" href="/">Gallery</a>
{{if .ID}}
<div id="display" class="visible">
  <div id="photo-side"><img id="live-photo" src="/photos/{{.PhotoPath}}" alt=""></div>
  <div id="info-side">
    <div id="live-score" class="s{{.WokeScore}}">{{.WokeScore}}</div>
    <div id="live-level" class="s{{.WokeScore}}">{{scoreLabel .WokeScore}}</div>
    <p id="live-desc">{{.Description}}</p>
    <div id="live-roast" style="display:none">
      <p id="live-roast-label" style="font-size:0.75rem;font-weight:700;letter-spacing:0.2em;text-transform:uppercase;margin-bottom:0.4rem"></p>
      <p id="live-roast-text" style="font-size:1rem;line-height:1.5;color:#888;max-width:480px;font-style:italic"></p>
    </div>
    <div>
      <div style="font-size:0.75rem;font-weight:700;letter-spacing:0.15em;text-transform:uppercase;color:#ccc;margin-bottom:0.3rem">WOOK↔WOKE Automatic Analysis ™</div>
      <div class="spectrum-wrap">
        <div class="spectrum-labels"><span class="lbl-wook">wook</span><span class="lbl-woke">woke</span></div>
        <div class="spectrum-track">
          <div id="spectrum-dot"{{if gt (scorePos .WokeScore) -1}} style="left:{{scorePos .WokeScore}}%"{{else}} style="display:none"{{end}}></div>
          <div id="spectrum-unknown"{{if eq (scorePos .WokeScore) -1}} style="display:flex"{{end}}><span>?</span></div>
        </div>
      </div>
    </div>
    <div id="live-vote">
      <div class="vote-label" style="color:#ccc;font-size:0.75rem">USER VOTES <span id="live-vstats">{{if gt .VoteCount 0}}Avg: {{printf1f .VoteAvg}} ({{.VoteCount}} {{if eq .VoteCount 1}}vote{{else}}votes{{end}}){{if gt .UserVote 0}} · Yours: {{.UserVote}}{{end}}{{else}}Be the first to vote!{{end}}</span></div>
      <div class="spectrum-wrap">
        <div class="spectrum-labels"><span class="lbl-wook">wook</span><span class="lbl-woke">woke</span></div>
        <div class="spectrum-track">
          <input type="range" min="1" max="7" step="1" class="vote-slider" id="live-vslider"
                 value="{{if gt .UserVote 0}}{{.UserVote}}{{else}}4{{end}}"
                 {{if eq .UserVote 0}}style="opacity:0.5"{{end}}
                 onchange="submitLiveVote(this)">
          <div class="vote-avg-marker" id="live-vmarker" style="{{if gt .VoteCount 0}}left:{{votePos .VoteAvg}}%{{else}}display:none{{end}}"></div>
        </div>
      </div>
    </div>
  </div>
</div>
{{else}}
<div id="waiting">
  <div style="font-size:3rem;font-weight:900"><span style="color:#f97316">WOOK</span><span style="color:#444;margin:0 0.3rem;font-weight:300">⟷</span><span style="color:#818cf8">WOKE</span></div>
  <h2>Waiting for first subject<span class="blink">_</span></h2>
</div>
{{end}}
<div id="analyzing" class="hidden">
  <div style="font-size:2.5rem;font-weight:900"><span style="color:#f97316">WOOK</span><span style="color:#444;margin:0 0.3rem;font-weight:300">⟷</span><span style="color:#818cf8">WOKE</span></div>
  <div class="scan-label">spectrum analysis in progress</div>
  <div id="analysis-msg">INITIALIZING<span class="blink">...</span></div>
  <div class="progress-track"><div id="progress-fill"></div></div>
</div>
<script>
  let currentId        = {{if .ID}}{{.ID}}{{else}}0{{end}};
  let currentPhoto     = {{if .PhotoPath}}'{{.PhotoPath}}'{{else}}''{{end}};
  let currentRescoreId = 0;

  function updateLiveVoteUI(avg, count, userVote) {
    const stats  = document.getElementById('live-vstats');
    const marker = document.getElementById('live-vmarker');
    const slider = document.getElementById('live-vslider');
    if (stats) {
      let txt = count > 0
        ? 'Avg: ' + avg.toFixed(1) + ' (' + count + (count === 1 ? ' vote' : ' votes') + ')'
        : 'Be the first to vote!';
      if (userVote > 0) txt += ' · Yours: ' + userVote;
      stats.textContent = txt;
    }
    if (marker) {
      if (count > 0) { marker.style.left = ((avg - 1) / 6 * 100).toFixed(1) + '%'; marker.style.display = ''; }
      else marker.style.display = 'none';
    }
    if (slider) {
      if (userVote > 0) { slider.value = userVote; slider.style.opacity = '1'; }
      else { slider.value = 4; slider.style.opacity = '0.4'; }
    }
  }

  async function submitLiveVote(slider) {
    if (!currentId) return;
    const score = parseInt(slider.value);
    try {
      const res = await fetch('/api/entries/' + currentId + '/vote', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ score }),
      });
      if (!res.ok) return;
      const data = await res.json();
      updateLiveVoteUI(data.avg, data.count, data.userVote);
    } catch (_) {}
  }

  const LEVELS = ['ESCAPED CLASSIFICATION','LEVEL 3 WOOK','LEVEL 2 WOOK','LEVEL 1 WOOK','NORMIE','LEVEL 1 WOKE','LEVEL 2 WOKE','LEVEL 3 WOKE'];

  const WOKE_ROASTS = [
    "You cancelled someone for a 2009 tweet, a Halloween costume from college, and vibes. The trial lasted four hours on a Discord server. You were judge, jury, and main character.",
    "You cancelled your best friend of eleven years over a repost and felt nothing. You posted about the grief though. Fourteen people liked it. You screenshotted that too.",
  ];

  const WOOK_ROASTS = [
    "You've been \"on the verge of awakening\" for eight straight festival seasons. You own 11 crystals and zero accountability. Every bad decision was \"a lesson.\"",
    "You don't live anywhere — you hover. You've paid rent in grilled cheese, half a tank hit, and a promise to \"get you next set.\" Your credit score is spiritual.",
    "You talk about \"escaping the system\" while day trading on venue WiFi. Your crystals are tax write-offs. You've pitched a DAO to someone actively on ketamine.",
    "You smell like sage and damp socks. You haven't showered but you have realigned your chakras in a mud puddle. You think deodorant blocks intuition.",
    "You met the DJ once in 2019 and won't shut up about it. You've traded grilled cheese for balloons and called it \"community exchange.\" You hug people mid-argument.",
    "You blame Mercury for things you absolutely did on purpose. You refuse to text back but will send a 14-minute voice memo about retrograde energy.",
    "You own nothing but somehow require everyone else's stuff. You say \"I don't believe in material attachment\" while asking to borrow socks.",
  ];
  const MESSAGES = [
    'SCANNING WOOK FREQUENCIES','DETECTING CRYSTAL ENERGY','CALIBRATING TIE-DYE DETECTOR',
    'MEASURING FESTIVAL RESIDUE','ANALYZING OUTFIT CHAOS','CROSS-REFERENCING DRUM CIRCLE DATABASE',
    'CHECKING FOR HIDDEN CRYSTALS','EVALUATING WOKE POTENTIAL','COMPUTING NORMIE COEFFICIENT',
    'ASSESSING GEOMETRIC PRECISION','CONSULTING THE SPIRIT GUIDES','TRIANGULATING CHAKRA ALIGNMENT',
    'SCANNING FOR PATCHOULI SIGNATURES','LOADING MUSEUM READINESS INDEX','QUERYING THE WOKE ORACLE',
  ];

  // Populate roast on initial page load
  (function() {
    const score = {{if .ID}}{{.WokeScore}}{{else}}-1{{end}};
    const el = document.getElementById('live-roast');
    if (el && score === 1) {
      document.getElementById('live-roast-label').textContent = '⚠ EXTREME WOOK DETECTED';
      document.getElementById('live-roast-label').style.color = '#fb923c';
      document.getElementById('live-roast-text').textContent = WOOK_ROASTS[Math.floor(Math.random() * WOOK_ROASTS.length)];
      el.style.display = '';
    } else if (el && score === 7) {
      document.getElementById('live-roast-label').textContent = '⚠ EXTREME WOKE DETECTED';
      document.getElementById('live-roast-label').style.color = '#c4b5fd';
      document.getElementById('live-roast-text').textContent = WOKE_ROASTS[Math.floor(Math.random() * WOKE_ROASTS.length)];
      el.style.display = '';
    }
  })();

  async function poll() {
    try {
      // Check for new entries
      const res = await fetch('/api/entries');
      if (!res.ok) return;
      const entries = await res.json();
      if (!entries || entries.length === 0) return;
      if (entries[0].ID > currentId) {
        currentId       = entries[0].ID;
        currentPhoto    = entries[0].PhotoPath;
        currentRescoreId = 0;
        await runAnalysis({ ...entries[0] });
        return; // skip rescore check this tick
      }

      // Update live vote stats from poll
      updateLiveVoteUI(entries[0].VoteAvg || 0, entries[0].VoteCount || 0, entries[0].UserVote || 0);

      // Check for new rescore on current entry — update in place, no animation
      if (currentId === 0) return;
      const rres = await fetch('/api/entries/' + currentId + '/rescores');
      if (!rres.ok) return;
      const rescores = await rres.json();
      if (!rescores || rescores.length === 0) return;
      const latest = rescores[0];
      if (latest.ID <= currentRescoreId) return;
      currentRescoreId = latest.ID;
      revealEntry({ WokeScore: latest.WokeScore, Description: latest.Description, PhotoPath: currentPhoto });
    } catch (_) {}
  }

  async function runAnalysis(entry) {
    const display  = document.getElementById('display');
    const waiting  = document.getElementById('waiting');
    const overlay  = document.getElementById('analyzing');
    const msgEl    = document.getElementById('analysis-msg');
    const fill     = document.getElementById('progress-fill');

    if (display) display.classList.remove('visible');
    if (waiting) waiting.style.display = 'none';
    overlay.classList.remove('hidden');
    fill.style.width = '0%';

    let msgIdx = Math.floor(Math.random() * MESSAGES.length);
    msgEl.innerHTML = MESSAGES[msgIdx] + '<span class="blink">...</span>';

    const duration = 3200;
    const start = Date.now();
    let lastMsg = start;
    await new Promise(resolve => {
      const tick = setInterval(() => {
        const elapsed = Date.now() - start;
        fill.style.width = Math.min(100, (elapsed / duration) * 100) + '%';
        if (Date.now() - lastMsg > 550) {
          msgIdx = (msgIdx + 1) % MESSAGES.length;
          msgEl.innerHTML = MESSAGES[msgIdx] + '<span class="blink">...</span>';
          lastMsg = Date.now();
        }
        if (elapsed >= duration) { clearInterval(tick); resolve(); }
      }, 60);
    });

    revealEntry(entry);
    overlay.classList.add('hidden');
    setTimeout(() => { if (display) display.classList.add('visible'); }, 50);
  }

  function revealEntry(entry) {
    const score = entry.WokeScore;
    const cls   = 's' + (score >= 0 && score <= 7 ? score : 4);
    const label = (score >= 0 && score <= 7) ? LEVELS[score] : 'OFF THE CHARTS';
    const pct   = score >= 1 && score <= 7 ? ((score - 1) / 6 * 100).toFixed(1) : null;

    let display = document.getElementById('display');
    if (!display) {
      display = document.createElement('div');
      display.id = 'display';
      display.innerHTML =
        '<div id="photo-side"><img id="live-photo" src="" alt=""></div>' +
        '<div id="info-side">' +
          '<div id="live-score"></div><div id="live-level"></div><p id="live-desc"></p>' +
          '<div id="live-roast" style="display:none"><p id="live-roast-label"></p><p id="live-roast-text"></p></div>' +
          '<div>' +
            '<div style="font-size:0.75rem;font-weight:700;letter-spacing:0.15em;text-transform:uppercase;color:#ccc;margin-bottom:0.3rem">WOOK↔WOKE Automatic Analysis ™</div>' +
            '<div class="spectrum-wrap">' +
              '<div class="spectrum-labels"><span class="lbl-wook">wook</span><span class="lbl-woke">woke</span></div>' +
              '<div class="spectrum-track"><div id="spectrum-dot"></div><div id="spectrum-unknown" style="display:none"><span>?</span></div></div>' +
            '</div>' +
          '</div>' +
          '<div id="live-vote">' +
            '<div class="vote-label" style="color:#ccc;font-size:0.75rem">USER VOTES <span id="live-vstats">Be the first to vote!</span></div>' +
            '<div class="spectrum-wrap">' +
              '<div class="spectrum-labels"><span class="lbl-wook">wook</span><span class="lbl-woke">woke</span></div>' +
              '<div class="spectrum-track">' +
                '<input type="range" min="1" max="7" step="1" class="vote-slider" id="live-vslider" value="4" style="opacity:0.5" onchange="submitLiveVote(this)">' +
                '<div class="vote-avg-marker" id="live-vmarker" style="display:none"></div>' +
              '</div>' +
            '</div>' +
          '</div>' +
        '</div>';
      document.body.appendChild(display);
    }
    // Reset vote section for new entry
    updateLiveVoteUI(entry.VoteAvg || 0, entry.VoteCount || 0, entry.UserVote || 0);

    document.getElementById('live-photo').src        = '/photos/' + entry.PhotoPath;
    const scoreEl = document.getElementById('live-score');
    scoreEl.textContent = score; scoreEl.className = cls;
    const levelEl = document.getElementById('live-level');
    levelEl.textContent = label; levelEl.className = cls;
    document.getElementById('live-desc').textContent  = entry.Description;
    const dot     = document.getElementById('spectrum-dot');
    const unknown = document.getElementById('spectrum-unknown');
    if (pct !== null) {
      dot.style.display = ''; dot.style.left = pct + '%';
      if (unknown) unknown.style.display = 'none';
    } else {
      dot.style.display = 'none';
      if (unknown) unknown.style.display = 'flex';
    }

    const roastEl = document.getElementById('live-roast');
    if (roastEl) {
      if (score === 1) {
        document.getElementById('live-roast-label').textContent = '⚠ EXTREME WOOK DETECTED';
        document.getElementById('live-roast-label').style.color = '#fb923c';
        document.getElementById('live-roast-text').textContent = WOOK_ROASTS[Math.floor(Math.random() * WOOK_ROASTS.length)];
        roastEl.style.display = '';
      } else if (score === 7) {
        document.getElementById('live-roast-label').textContent = '⚠ EXTREME WOKE DETECTED';
        document.getElementById('live-roast-label').style.color = '#c4b5fd';
        document.getElementById('live-roast-text').textContent = WOKE_ROASTS[Math.floor(Math.random() * WOKE_ROASTS.length)];
        roastEl.style.display = '';
      } else {
        roastEl.style.display = 'none';
      }
    }
  }

  setInterval(poll, 5000);
</script>
</body></html>`))

var galleryTmpl = template.Must(template.New("gallery").Funcs(tmplFuncs).Parse(`<!DOCTYPE html>
<html><head>
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>wook2woke</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: system-ui; background: #0a0a0a; color: #e0e0e0; padding: 1rem; }
  /* Header */
  header { text-align: center; margin-bottom: 0.5rem; }
  .logout-btn { font-size: 0.8rem; letter-spacing: 0.12em; text-transform: uppercase; color: #555; background: none; border: 1px solid #2a2a2a; border-radius: 6px; padding: 0.3rem 0.9rem; cursor: pointer; text-decoration: none; }
  .logout-btn:hover { color: #ccc; border-color: #666; }
  .header-title { font-size: 2.4rem; font-weight: 900; letter-spacing: -0.02em; line-height: 1; }
  .header-title .wook { color: #f97316; }
  .header-title .sep { color: #444; margin: 0 0.3rem; font-weight: 300; }
  .header-title .woke { color: #818cf8; }
  .header-bar { display: flex; align-items: center; justify-content: center; gap: 0.5rem; margin: 0.5rem auto 1.25rem; max-width: 360px; }
  .header-bar .bar { flex: 1; height: 3px; border-radius: 2px; background: linear-gradient(90deg, #f97316, #fbbf24, #6b7280, #60a5fa, #818cf8); }
  .header-bar .bar-label { font-size: 0.7rem; letter-spacing: 0.12em; text-transform: uppercase; color: #555; }
  /* Controls */
  .controls { display: flex; flex-wrap: wrap; justify-content: center; gap: 0.5rem; margin-bottom: 1.5rem; }
  .controls button { padding: 0.4rem 1rem; border: 1px solid #333; border-radius: 6px; background: #1a1a1a; color: #e0e0e0; font-size: 0.9rem; cursor: pointer; }
  .controls button.active { background: #4f46e5; border-color: #4f46e5; }
  /* Grid */
  .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 1rem; max-width: 1200px; margin: 0 auto; }
  .card { background: #1a1a1a; border-radius: 12px; overflow: hidden; position: relative; }
  .card-clickable { cursor: pointer; position: relative; border-radius: 12px 12px 0 0; }
  @media (hover: hover) {
    .card-clickable { transition: transform 0.2s ease, box-shadow 0.2s ease; }
    .card-clickable:hover { transform: translateY(-2px); box-shadow: 0 8px 24px rgba(0,0,0,0.5), 0 2px 8px rgba(0,0,0,0.3); }
  }
  .card img { width: 100%; aspect-ratio: 4/3; object-fit: cover; display: block; }
  .card .info { padding: 1rem 1rem 1.1rem; }
  .card .score-row { display: flex; align-items: center; gap: 1rem; flex-wrap: wrap; }
  .card .info .level { font-size: 0.85rem; padding: 0.35rem 0.75rem; }
  .card .info .score { font-size: 1.1rem; }
  /* Score number colors by data-score */
  .score { font-size: 2rem; font-weight: bold; }
  .score[data-score="0"] { color: #9ca3af; }
  .score[data-score="1"] { color: #c2410c; }
  .score[data-score="2"] { color: #f97316; }
  .score[data-score="3"] { color: #fbbf24; }
  .score[data-score="4"] { color: #d1d5db; }
  .score[data-score="5"] { color: #38bdf8; }
  .score[data-score="6"] { color: #60a5fa; }
  .score[data-score="7"] { color: #a78bfa; }
  /* Level badge colors by data-score */
  .level { font-size: 0.72rem; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase; border: 1px solid; border-radius: 4px; padding: 0.15rem 0.4rem; cursor: default; white-space: nowrap; }
  .level[data-score="0"] { color: #9ca3af; border-color: #4b5563; background: #1c1f26; }
  .level[data-score="1"] { color: #fb923c; border-color: #9a3412; background: #2c1008; }
  .level[data-score="2"] { color: #fbbf24; border-color: #b45309; background: #2c1a04; }
  .level[data-score="3"] { color: #fde68a; border-color: #ca8a04; background: #2a1f03; }
  .level[data-score="4"] { color: #d1d5db; border-color: #6b7280; background: #1f2937; }
  .level[data-score="5"] { color: #7dd3fc; border-color: #0369a1; background: #082030; }
  .level[data-score="6"] { color: #93c5fd; border-color: #1d4ed8; background: #0d1f42; }
  .level[data-score="7"] { color: #c4b5fd; border-color: #6d28d9; background: #1e0a40; }
  .card .desc { margin-top: 0.5rem; line-height: 1.4; }
  .card .time { margin-top: 0.5rem; font-size: 0.8rem; color: #666; }
  /* Vote slider */
  .vote-section { padding: 0.6rem 1rem 0.85rem; border-top: 1px solid #222; }
  .vote-header { display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 0.3rem; }
  .vote-title { font-size: 0.6rem; font-weight: 700; letter-spacing: 0.15em; text-transform: uppercase; color: #555; }
  .vote-stats { font-size: 0.7rem; color: #666; }
  .vote-bar-wrap { position: relative; }
  .vote-slider { -webkit-appearance: none; appearance: none; width: 100%; height: 64px; background: transparent; outline: none; cursor: pointer; display: block; transition: opacity 0.2s; margin: 0; }
  .vote-slider::-webkit-slider-runnable-track { height: 6px; border-radius: 3px; background: linear-gradient(90deg, #f97316, #fbbf24, #6b7280, #60a5fa, #818cf8); }
  .vote-slider::-moz-range-track { height: 6px; border-radius: 3px; background: linear-gradient(90deg, #f97316, #fbbf24, #6b7280, #60a5fa, #818cf8); }
  .vote-slider::-webkit-slider-thumb { -webkit-appearance: none; width: 28px; height: 28px; border-radius: 50%; background: white; cursor: pointer; border: 2px solid #333; box-shadow: 0 0 8px rgba(255,255,255,0.5); margin-top: -11px; }
  .vote-slider::-moz-range-thumb { width: 28px; height: 28px; border-radius: 50%; background: white; cursor: pointer; border: 2px solid #333; }
  .vote-avg-marker { position: absolute; top: 50%; transform: translate(-50%, -50%); width: 10px; height: 10px; border-radius: 50%; background: rgba(255,255,255,0.35); border: 1.5px solid rgba(255,255,255,0.7); pointer-events: none; transition: left 0.4s; }
  .vote-wook-woke { display: flex; justify-content: space-between; font-size: 0.6rem; letter-spacing: 0.1em; text-transform: uppercase; margin-top: 0.15rem; }
  .vote-wook-woke .lw { color: #f97316; }
  .vote-wook-woke .lk { color: #818cf8; }
  .card .rescore-badge { position: absolute; top: 0.5rem; left: 0.5rem; background: rgba(79,70,229,0.85); color: white; font-size: 0.75rem; padding: 0.2rem 0.5rem; border-radius: 999px; display: none; }
  .card.has-rescores .rescore-badge { display: block; }
  .card-hover { position: absolute; inset: 0; background: none; display: flex; align-items: flex-start; justify-content: space-between; padding: 0.5rem; opacity: 0; transition: opacity 0.15s; pointer-events: none; }
  .card-hover-right { display: flex; gap: 0.4rem; align-items: center; }
  .card-clickable:hover .card-hover { opacity: 1; pointer-events: auto; }
  .rescore-btn { padding: 0.4rem 0.8rem; background: #4f46e5; border: none; border-radius: 8px; color: white; font-size: 0.85rem; cursor: pointer; }
  .rescore-btn:hover { background: #4338ca; }
  .rescore-btn:disabled { background: #555; cursor: not-allowed; }
  .download-btn { padding: 0.4rem 0.8rem; background: rgba(0,0,0,0.6); border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; color: white; font-size: 0.85rem; cursor: pointer; text-decoration: none; line-height: 1.4; }
  .delete-btn { padding: 0.4rem 0.6rem; background: rgba(127,29,29,0.7); border: 1px solid rgba(239,68,68,0.4); border-radius: 8px; color: white; font-size: 0.85rem; cursor: pointer; }
  /* Modal */
  .modal { display: none; position: fixed; inset: 0; z-index: 100; }
  .modal.open { display: flex; align-items: center; justify-content: center; }
  .modal-backdrop { position: absolute; inset: 0; background: rgba(0,0,0,0.75); }
  .modal-box { position: relative; background: #1a1a1a; border-radius: 12px; width: 92%; max-width: 560px; max-height: 90vh; overflow-y: auto; z-index: 1; }
  .modal-close { position: absolute; top: 0.75rem; right: 0.75rem; background: none; border: none; color: #aaa; font-size: 1.3rem; cursor: pointer; line-height: 1; }
  .modal-orig { display: flex; gap: 1rem; padding: 1.25rem; border-bottom: 1px solid #2a2a2a; align-items: flex-start; }
  .modal-orig img { width: 120px; height: 90px; object-fit: cover; border-radius: 8px; flex-shrink: 0; }
  .modal-orig .score-row { display: flex; gap: 1rem; align-items: stretch; }
  .modal-orig .level { font-size: 0.85rem; padding: 0.35rem 0.75rem; display: inline-flex; align-items: center; }
  .modal-orig .score { font-size: 1.1rem; }
  .modal-rescores { padding: 1rem 1.25rem; }
  .modal-rescores h3 { margin-bottom: 0.75rem; font-size: 0.9rem; color: #888; text-transform: uppercase; letter-spacing: 0.05em; }
  .rs-card { display: flex; gap: 1rem; align-items: flex-start; padding: 0.75rem; background: #111; border-radius: 8px; margin-bottom: 0.5rem; }
  .rs-card img { width: 80px; height: 60px; object-fit: cover; border-radius: 6px; flex-shrink: 0; }
  .rs-card .score { font-size: 1.1rem; }
  .rs-card .level { display: inline-flex; align-items: center; }
  .rs-card .rs-subject { font-size: 0.75rem; color: #666; margin-top: 0.5rem; margin-bottom: 0.2rem; }
  .rs-card .rs-desc { font-size: 0.9rem; line-height: 1.3; }
  .rs-badge { display: inline-block; font-size: 0.7rem; padding: 0.15rem 0.4rem; background: #312e81; color: #a5b4fc; border-radius: 4px; margin-bottom: 0.3rem; }
  .promote-btn { margin-top: 0.5rem; padding: 0.3rem 0.7rem; background: #14532d; border: 1px solid #16a34a; border-radius: 6px; color: #86efac; font-size: 0.8rem; cursor: pointer; }
  .empty { text-align: center; color: #666; margin-top: 4rem; }
</style>
</head><body>
<header>
  <div class="header-title">
    <span class="wook">WOOK</span><span class="sep">⟷</span><span class="woke">WOKE</span>
  </div>
  <div class="header-bar">
    <span class="bar-label">wook</span>
    <div class="bar"></div>
    <span class="bar-label">woke</span>
  </div>
  <div style="text-align:center;margin-bottom:1.5rem;display:flex;justify-content:center;gap:0.6rem;">
    <a href="/live" style="font-size:0.8rem;letter-spacing:0.12em;text-transform:uppercase;color:#555;text-decoration:none;border:1px solid #2a2a2a;border-radius:6px;padding:0.3rem 0.9rem;">⬤ Live View</a>
    {{if .IsAdmin}}<form action="/logout" method="post" style="margin:0"><button class="logout-btn" type="submit">Logout</button></form>{{else}}<a href="/login" class="logout-btn">Admin Login</a>{{end}}
  </div>
</header>
<div class="controls">
  <button onclick="setSort('newest')" id="btn-newest" class="active">Most Recent</button>
  <button onclick="setSort('oldest')" id="btn-oldest">First Victims</button>
  <button onclick="setSort('mostwook')" id="btn-mostwook">Smelliest</button>
  <button onclick="setSort('mostwoke')" id="btn-mostwoke">Most Judgmental</button>
  <button onclick="setSort('normies')" id="btn-normies">Normies</button>
  <button onclick="setSort('unknown')" id="btn-unknown">Escaped Classification</button>
</div>
<!-- Rescore modal -->
<div class="modal" id="modal">
  <div class="modal-backdrop" onclick="closeModal()"></div>
  <div class="modal-box">
    <button class="modal-close" onclick="closeModal()">✕</button>
    <div id="modal-body"></div>
  </div>
</div>

{{if .Entries}}
<div class="grid" id="grid">
  {{range .Entries}}
  <div class="card" data-id="{{.ID}}" data-score="{{.WokeScore}}" data-time="{{.CreatedAt}}" data-photo="{{.PhotoPath}}" data-desc="{{.Description}}">
    <div class="card-clickable" onclick="openModal(this.closest('.card'))">
    <span class="rescore-badge" id="badge-{{.ID}}">✨ re-analyzed</span>
    <img src="/photos/{{.PhotoPath}}" alt="entry {{.ID}}" loading="lazy">
    <div class="card-hover">
      {{if $.IsAdmin}}
      <button class="rescore-btn" onclick="event.stopPropagation(); triggerRescore(this, {{.ID}})">✨ Rescore</button>
      <div class="card-hover-right">
        <button class="delete-btn" onclick="event.stopPropagation(); deleteEntry(this, {{.ID}})">🗑</button>
        <a class="download-btn" href="/photos/{{.PhotoPath}}" download onclick="event.stopPropagation()">⬇ Save</a>
      </div>
      {{else}}
      <div></div>
      <a class="download-btn" href="/photos/{{.PhotoPath}}" download onclick="event.stopPropagation()">⬇ Save</a>
      {{end}}
    </div>
    <div class="info">
      <div class="score-row">
        <span class="level" data-score="{{.WokeScore}}" title="{{scoreVibe .WokeScore}}">{{scoreLabel .WokeScore}}</span>
        <span class="score" data-score="{{.WokeScore}}">{{.WokeScore}}</span>
      </div>
      <p class="desc">{{.Description}}</p>
      <p class="time">{{.CreatedAt}}</p>
    </div>
    </div>
    <div class="vote-section">
      <div class="vote-header">
        <span class="vote-title">USER VOTES</span>
        <span class="vote-stats" id="vstats-{{.ID}}">{{if gt .VoteCount 0}}Avg: {{printf1f .VoteAvg}} ({{.VoteCount}} {{if eq .VoteCount 1}}vote{{else}}votes{{end}}){{if gt .UserVote 0}} · Yours: {{.UserVote}}{{end}}{{else}}Be the first to vote!{{end}}</span>
      </div>
      <div class="vote-bar-wrap">
        <input type="range" min="1" max="7" step="1" class="vote-slider" id="vslider-{{.ID}}"
               value="{{if gt .UserVote 0}}{{.UserVote}}{{else}}4{{end}}"
               {{if eq .UserVote 0}}style="opacity:0.4"{{end}}
               onchange="submitVote(this,{{.ID}})">
        <div class="vote-avg-marker" id="vmarker-{{.ID}}" style="{{if gt .VoteCount 0}}left:{{votePos .VoteAvg}}%{{else}}display:none{{end}}"></div>
      </div>
      <div class="vote-wook-woke"><span class="lw">wook</span><span class="lk">woke</span></div>
    </div>
  </div>
  {{end}}
</div>
{{else}}
<p class="empty" id="empty-msg">No entries yet. Waiting for the ESP to send some photos...</p>
{{end}}
<script>
  const isAdmin = {{if .IsAdmin}}true{{else}}false{{end}};
  let latestId = {{if .Entries}}{{(index .Entries 0).ID}}{{else}}0{{end}};
  let currentSort = 'newest';

  const LEVELS = [
    { label: 'ESCAPED CLASSIFICATION', vibe: 'Escaped classification entirely' },
    { label: 'LEVEL 3 WOOK',  vibe: 'Raw, earthy, chaotic, festival energy' },
    { label: 'LEVEL 2 WOOK',  vibe: 'Warm, cloudy, irregular, rough' },
    { label: 'LEVEL 1 WOOK',  vibe: 'Leaning wook' },
    { label: 'NORMIE',        vibe: 'Dead center, neither wook nor woke' },
    { label: 'LEVEL 1 WOKE',  vibe: 'Leaning woke' },
    { label: 'LEVEL 2 WOKE',  vibe: 'Cool, clear, geometric, polished' },
    { label: 'LEVEL 3 WOKE',  vibe: 'Polished, geometric, precise, museum-ready' },
  ];

  function levelTagLarge(score) {
    const s = parseInt(score);
    const style = 'font-size:0.85rem;padding:0 0.7rem;display:inline-flex;align-items:center;';
    if (s < 0 || s > 7) return '<span class="level" data-score="-1" style="' + style + '" title="Score outside the 0–7 range">OFF THE CHARTS</span>';
    return '<span class="level" data-score="' + s + '" style="' + style + '" title="' + LEVELS[s].vibe + '">' + LEVELS[s].label + '</span>';
  }

  function levelTag(score) {
    const s = parseInt(score);
    if (s < 0 || s > 7) return '<span class="level" data-score="-1" title="Score outside the 0–7 range">OFF THE CHARTS</span>';
    return '<span class="level" data-score="' + s + '" title="' + LEVELS[s].vibe + '">' + LEVELS[s].label + '</span>';
  }

  function scoreTag(score) {
    return '<span class="score" data-score="' + score + '">' + score + '</span>';
  }

  function makeVoteSection(e) {
    const hasVote  = e.UserVote > 0;
    const hasVotes = e.VoteCount > 0;
    let stats = hasVotes
      ? 'Avg: ' + e.VoteAvg.toFixed(1) + ' (' + e.VoteCount + (e.VoteCount === 1 ? ' vote' : ' votes') + ')'
      : 'Be the first to vote!';
    if (hasVote) stats += ' · Yours: ' + e.UserVote;
    const pct = hasVotes ? ((e.VoteAvg - 1) / 6 * 100).toFixed(1) : 0;
    return '<div class="vote-section" onclick="event.stopPropagation()">' +
      '<div class="vote-header">' +
        '<span class="vote-title">USER VOTES</span>' +
        '<span class="vote-stats" id="vstats-' + e.ID + '">' + stats + '</span>' +
      '</div>' +
      '<div class="vote-bar-wrap">' +
        '<input type="range" min="1" max="7" step="1" class="vote-slider" id="vslider-' + e.ID + '" ' +
          'value="' + (hasVote ? e.UserVote : 4) + '" ' +
          'style="' + (hasVote ? '' : 'opacity:0.4') + '" ' +
          'onchange="submitVote(this,' + e.ID + ')">' +
        '<div class="vote-avg-marker" id="vmarker-' + e.ID + '" style="' + (hasVotes ? 'left:' + pct + '%' : 'display:none') + '"></div>' +
      '</div>' +
      '<div class="vote-wook-woke"><span class="lw">wook</span><span class="lk">woke</span></div>' +
    '</div>';
  }

  function makeCard(e) {
    const hoverContent = isAdmin
      ? '<button class="rescore-btn" onclick="event.stopPropagation(); triggerRescore(this,' + e.ID + ')">✨ Rescore</button>' +
        '<div class="card-hover-right"><button class="delete-btn" onclick="event.stopPropagation(); deleteEntry(this,' + e.ID + ')">🗑</button><a class="download-btn" href="/photos/' + e.PhotoPath + '" download onclick="event.stopPropagation()">⬇ Save</a></div>'
      : '<div></div><a class="download-btn" href="/photos/' + e.PhotoPath + '" download onclick="event.stopPropagation()">⬇ Save</a>';
    return '<div class="card" data-id="' + e.ID + '" data-score="' + e.WokeScore + '" data-time="' + e.CreatedAt + '" data-photo="' + e.PhotoPath + '" data-desc="' + e.Description + '">' +
      '<div class="card-clickable" onclick="openModal(this.closest(\'.card\'))">' +
        '<span class="rescore-badge" id="badge-' + e.ID + '">✨ re-analyzed</span>' +
        '<img src="/photos/' + e.PhotoPath + '" alt="entry ' + e.ID + '" loading="lazy">' +
        '<div class="card-hover">' + hoverContent + '</div>' +
        '<div class="info"><div class="score-row">' + levelTag(e.WokeScore) + scoreTag(e.WokeScore) + '</div><p class="desc">' + e.Description + '</p><p class="time">' + e.CreatedAt + '</p></div>' +
      '</div>' +
      makeVoteSection(e) +
      '</div>';
  }

  function setSort(mode) {
    currentSort = mode;
    document.querySelectorAll('.controls button').forEach(b => b.classList.remove('active'));
    const btn = document.getElementById('btn-' + mode);
    if (btn) btn.classList.add('active');
    sortGrid();
  }

  function sortGrid() {
    const grid = document.getElementById('grid');
    if (!grid) return;
    const allCards = Array.from(grid.querySelectorAll('.card'));

    allCards.forEach(c => c.style.display = '');

    if (currentSort === 'normies') {
      allCards.forEach(c => {
        c.style.display = Number(c.dataset.score) === 4 ? '' : 'none';
      });
      return;
    }

    if (currentSort === 'unknown') {
      allCards.forEach(c => {
        c.style.display = Number(c.dataset.score) === 0 ? '' : 'none';
      });
      return;
    }

    const cards = allCards.filter(c => c.style.display !== 'none');
    cards.sort((a, b) => {
      if (currentSort === 'mostwook') {
        const sa = Number(a.dataset.score), sb = Number(b.dataset.score);
        return (sa === 0 ? Infinity : sa) - (sb === 0 ? Infinity : sb);
      }
      if (currentSort === 'mostwoke') {
        const sa = Number(a.dataset.score), sb = Number(b.dataset.score);
        return (sb === 0 ? -Infinity : sb) - (sa === 0 ? -Infinity : sa);
      }
      if (currentSort === 'oldest')    return a.dataset.time.localeCompare(b.dataset.time);
      return b.dataset.time.localeCompare(a.dataset.time); // newest
    });
    cards.forEach(c => grid.appendChild(c));
  }

  function updateVoteUI(entryId, avg, count, userVote) {
    const stats  = document.getElementById('vstats-' + entryId);
    const marker = document.getElementById('vmarker-' + entryId);
    const slider = document.getElementById('vslider-' + entryId);
    if (stats) {
      let txt = count > 0
        ? 'Avg: ' + avg.toFixed(1) + ' (' + count + (count === 1 ? ' vote' : ' votes') + ')'
        : 'Be the first to vote!';
      if (userVote > 0) txt += ' · Yours: ' + userVote;
      stats.textContent = txt;
    }
    if (marker) {
      if (count > 0) {
        const pct = ((avg - 1) / 6 * 100).toFixed(1);
        marker.style.left = pct + '%';
        marker.style.display = '';
      } else {
        marker.style.display = 'none';
      }
    }
    if (slider && userVote > 0) {
      slider.value = userVote;
      slider.style.opacity = '1';
    }
  }

  async function submitVote(slider, entryId) {
    const score = parseInt(slider.value);
    try {
      const res = await fetch('/api/entries/' + entryId + '/vote', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ score }),
      });
      if (!res.ok) return;
      const data = await res.json();
      updateVoteUI(entryId, data.avg, data.count, data.userVote);
    } catch (_) {}
  }

  async function poll() {
    try {
      const res = await fetch('/api/entries');
      if (!res.ok) return;
      const entries = await res.json();
      if (!entries || entries.length === 0) return;

      // Update vote stats on all existing cards
      entries.forEach(e => updateVoteUI(e.ID, e.VoteAvg, e.VoteCount, e.UserVote));

      const newEntries = entries.filter(e => e.ID > latestId);
      if (newEntries.length === 0) return;

      latestId = entries[0].ID;

      const emptyMsg = document.getElementById('empty-msg');
      if (emptyMsg) emptyMsg.remove();

      let grid = document.getElementById('grid');
      if (!grid) {
        grid = document.createElement('div');
        grid.className = 'grid';
        grid.id = 'grid';
        document.querySelector('.controls').insertAdjacentElement('afterend', grid);
      }

      newEntries.forEach(e => grid.insertAdjacentHTML('afterbegin', makeCard(e)));
      sortGrid();
    } catch (_) {}
  }

  setInterval(poll, 5000);

  // --- Rescore ---

  async function triggerRescore(btn, entryId) {
    btn.disabled = true;
    btn.textContent = 'Analyzing...';
    try {
      const res = await fetch('/api/entries/' + entryId + '/rescore', { method: 'POST' });
      if (!res.ok) { btn.textContent = 'Error'; return; }
      const rs = await res.json();
      const card = document.querySelector('.card[data-id="' + entryId + '"]');
      if (card) card.classList.add('has-rescores');
      btn.textContent = '✨ Rescore';
    } catch (_) {
      btn.textContent = 'Error';
    } finally {
      btn.disabled = false;
    }
  }

  async function promoteRescore(entryId, score, description) {
    try {
      const res = await fetch('/api/entries/' + entryId, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ wokeScore: score, description }),
      });
      if (!res.ok) { alert('Promote failed'); return; }
      // Update the card in the gallery
      const card = document.querySelector('.card[data-id="' + entryId + '"]');
      if (card) {
        card.dataset.score = score;
        card.dataset.desc = description;
        card.querySelector('.score').textContent = score;
        card.querySelector('.score').dataset.score = score;
        card.querySelector('.level').textContent = (LEVELS[score] && LEVELS[score].label) || 'OFF THE CHARTS';
        card.querySelector('.level').dataset.score = score;
        card.querySelector('.desc').textContent = description;
      }
      closeModal();
    } catch (_) { alert('Promote failed'); }
  }

  async function deleteEntry(btn, entryId) {
    if (!confirm('Delete this entry?')) return;
    try {
      const res = await fetch('/api/entries/' + entryId, { method: 'DELETE' });
      if (!res.ok) { alert('Delete failed'); return; }
      document.querySelector('.card[data-id="' + entryId + '"]')?.remove();
    } catch (_) { alert('Delete failed'); }
  }

  // --- Modal ---

  async function openModal(card) {
    const entryId = card.dataset.id;
    const photo = card.dataset.photo;
    const score = card.dataset.score;
    const desc = card.dataset.desc;
    const time = card.querySelector('.time').textContent;

    document.getElementById('modal-body').innerHTML =
      '<div class="modal-orig">' +
        '<img src="/photos/' + photo + '" alt="">' +
        '<div><div class="score-row" style="gap:1rem;align-items:stretch">' + levelTagLarge(score) + scoreTag(score) + '</div><p style="margin-top:0.3rem">' + desc + '</p><p style="font-size:0.8rem;color:#666;margin-top:0.3rem">' + time + '</p></div>' +
      '</div>' +
      '<div class="modal-rescores"><h3>Re-analyses</h3><div id="rs-list"><p style="color:#666;font-size:0.9rem">Loading...</p></div></div>';

    document.getElementById('modal').classList.add('open');

    try {
      const res = await fetch('/api/entries/' + entryId + '/rescores');
      const rescores = await res.json();
      const list = document.getElementById('rs-list');
      if (!rescores || rescores.length === 0) {
        list.innerHTML = isAdmin
          ? '<p style="color:#666;font-size:0.9rem">No re-analyses yet. Hover the card and click ✨ Rescore.</p>'
          : '<p style="color:#666;font-size:0.9rem">Give Jordan or Calvin a beer for a rescore 😉</p>';
        return;
      }
      const card = document.querySelector('.card[data-id="' + entryId + '"]');
      if (card) card.classList.add('has-rescores');
      list.innerHTML = rescores.map(rs =>
        '<div class="rs-card">' +
          '<img src="/photos/' + photo + '" alt="">' +
          '<div style="flex:1"><span class="rs-badge">✨ re-analyzed</span><div class="score-row" style="display:flex;align-items:stretch;gap:1rem">' + levelTag(rs.WokeScore) + scoreTag(rs.WokeScore) + '</div><div class="rs-subject">' + rs.Subject + '</div><p class="rs-desc">' + rs.Description + '</p>' +
          (isAdmin ? '<button class="promote-btn" onclick="promoteRescore(' + entryId + ',' + rs.WokeScore + ',\'' + rs.Description.replace(/'/g,"\\\'") + '\')">↑ Promote to official score</button>' : '') +
          '</div>' +
        '</div>'
      ).join('');
    } catch (_) {
      document.getElementById('rs-list').innerHTML = '<p style="color:#ef4444">Failed to load.</p>';
    }
  }

  function closeModal() {
    document.getElementById('modal').classList.remove('open');
  }

  document.addEventListener('keydown', e => { if (e.key === 'Escape') closeModal(); });

  // On touch devices, open modal directly on tap without requiring a hover state first.
  // Only fire if the finger didn't scroll (movement < 10px).
  if ('ontouchstart' in window) {
    let touchStartX = 0, touchStartY = 0;
    const grid = document.getElementById('grid');
    grid?.addEventListener('touchstart', e => {
      touchStartX = e.touches[0].clientX;
      touchStartY = e.touches[0].clientY;
    }, { passive: true });
    grid?.addEventListener('touchend', e => {
      const dx = Math.abs(e.changedTouches[0].clientX - touchStartX);
      const dy = Math.abs(e.changedTouches[0].clientY - touchStartY);
      if (dx > 10 || dy > 10) return; // was a scroll, not a tap
      const clickable = e.target.closest('.card-clickable');
      if (clickable) {
        e.preventDefault();
        openModal(clickable.closest('.card'));
      }
    }, { passive: false });
  }
</script>
</body></html>`))
