package server

import "html/template"

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

var galleryTmpl = template.Must(template.New("gallery").Parse(`<!DOCTYPE html>
<html><head>
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>wook2woke</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: system-ui; background: #0a0a0a; color: #e0e0e0; padding: 1rem; }
  h1 { text-align: center; margin-bottom: 1rem; font-size: 1.8rem; }
  .controls { display: flex; justify-content: center; gap: 0.5rem; margin-bottom: 1.5rem; }
  .controls button { padding: 0.4rem 1rem; border: 1px solid #333; border-radius: 6px; background: #1a1a1a; color: #e0e0e0; font-size: 0.9rem; cursor: pointer; }
  .controls button.active { background: #4f46e5; border-color: #4f46e5; }
  .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 1rem; max-width: 1200px; margin: 0 auto; }
  .card { background: #1a1a1a; border-radius: 12px; overflow: hidden; cursor: pointer; position: relative; }
  .card img { width: 100%; aspect-ratio: 4/3; object-fit: cover; display: block; }
  .card .info { padding: 1rem; }
  .card .score { font-size: 2rem; font-weight: bold; color: #4f46e5; }
  .card .desc { margin-top: 0.5rem; line-height: 1.4; }
  .card .time { margin-top: 0.5rem; font-size: 0.8rem; color: #666; }
  .card .rescore-badge { position: absolute; top: 0.5rem; right: 0.5rem; background: rgba(79,70,229,0.85); color: white; font-size: 0.75rem; padding: 0.2rem 0.5rem; border-radius: 999px; display: none; }
  .card.has-rescores .rescore-badge { display: block; }
  .card-hover { position: absolute; inset: 0; background: rgba(0,0,0,0.55); display: flex; align-items: center; justify-content: center; opacity: 0; transition: opacity 0.15s; pointer-events: none; }
  .card:hover .card-hover { opacity: 1; pointer-events: auto; }
  .rescore-btn { padding: 0.6rem 1.2rem; background: #4f46e5; border: none; border-radius: 8px; color: white; font-size: 0.95rem; cursor: pointer; }
  .rescore-btn:hover { background: #4338ca; }
  .rescore-btn:disabled { background: #555; cursor: not-allowed; }
  /* Modal */
  .modal { display: none; position: fixed; inset: 0; z-index: 100; }
  .modal.open { display: flex; align-items: center; justify-content: center; }
  .modal-backdrop { position: absolute; inset: 0; background: rgba(0,0,0,0.75); }
  .modal-box { position: relative; background: #1a1a1a; border-radius: 12px; width: 92%; max-width: 560px; max-height: 90vh; overflow-y: auto; z-index: 1; }
  .modal-close { position: absolute; top: 0.75rem; right: 0.75rem; background: none; border: none; color: #aaa; font-size: 1.3rem; cursor: pointer; line-height: 1; }
  .modal-orig { display: flex; gap: 1rem; padding: 1.25rem; border-bottom: 1px solid #2a2a2a; align-items: flex-start; }
  .modal-orig img { width: 120px; height: 90px; object-fit: cover; border-radius: 8px; flex-shrink: 0; }
  .modal-orig .score { font-size: 1.8rem; font-weight: bold; color: #4f46e5; }
  .modal-rescores { padding: 1rem 1.25rem; }
  .modal-rescores h3 { margin-bottom: 0.75rem; font-size: 0.9rem; color: #888; text-transform: uppercase; letter-spacing: 0.05em; }
  .rs-card { display: flex; gap: 1rem; align-items: flex-start; padding: 0.75rem; background: #111; border-radius: 8px; margin-bottom: 0.5rem; }
  .rs-card img { width: 80px; height: 60px; object-fit: cover; border-radius: 6px; flex-shrink: 0; }
  .rs-card .rs-score { font-size: 1.4rem; font-weight: bold; color: #818cf8; }
  .rs-card .rs-subject { font-size: 0.75rem; color: #666; margin-bottom: 0.2rem; }
  .rs-card .rs-desc { font-size: 0.9rem; line-height: 1.3; }
  .rs-badge { display: inline-block; font-size: 0.7rem; padding: 0.15rem 0.4rem; background: #312e81; color: #a5b4fc; border-radius: 4px; margin-bottom: 0.3rem; }
  .empty { text-align: center; color: #666; margin-top: 4rem; }
</style>
</head><body>
<h1>wook2woke</h1>
<div class="controls">
  <button onclick="setSort('newest')" id="btn-newest" class="active">Newest</button>
  <button onclick="setSort('oldest')" id="btn-oldest">Oldest</button>
  <button onclick="setSort('score')" id="btn-score">Top Score</button>
</div>
<!-- Rescore modal -->
<div class="modal" id="modal">
  <div class="modal-backdrop" onclick="closeModal()"></div>
  <div class="modal-box">
    <button class="modal-close" onclick="closeModal()">✕</button>
    <div id="modal-body"></div>
  </div>
</div>

{{if .}}
<div class="grid" id="grid">
  {{range .}}
  <div class="card" data-id="{{.ID}}" data-score="{{.WokeScore}}" data-time="{{.CreatedAt}}" data-photo="{{.PhotoPath}}" data-desc="{{.Description}}" onclick="openModal(this)">
    <span class="rescore-badge" id="badge-{{.ID}}">✨ re-analyzed</span>
    <img src="/photos/{{.PhotoPath}}" alt="entry {{.ID}}" loading="lazy">
    <div class="card-hover">
      <button class="rescore-btn" onclick="event.stopPropagation(); triggerRescore(this, {{.ID}})">✨ Rescore</button>
    </div>
    <div class="info">
      <span class="score">{{.WokeScore}}</span>
      <p class="desc">{{.Description}}</p>
      <p class="time">{{.CreatedAt}}</p>
    </div>
  </div>
  {{end}}
</div>
{{else}}
<p class="empty" id="empty-msg">No entries yet. Waiting for the ESP to send some photos...</p>
{{end}}
<script>
  let latestId = {{if .}}{{(index . 0).ID}}{{else}}0{{end}};
  let currentSort = 'newest';

  function makeCard(e) {
    return '<div class="card" data-id="' + e.ID + '" data-score="' + e.WokeScore + '" data-time="' + e.CreatedAt + '" data-photo="' + e.PhotoPath + '" data-desc="' + e.Description + '" onclick="openModal(this)">' +
      '<span class="rescore-badge" id="badge-' + e.ID + '">✨ re-analyzed</span>' +
      '<img src="/photos/' + e.PhotoPath + '" alt="entry ' + e.ID + '" loading="lazy">' +
      '<div class="card-hover"><button class="rescore-btn" onclick="event.stopPropagation(); triggerRescore(this, ' + e.ID + ')">✨ Rescore</button></div>' +
      '<div class="info"><span class="score">' + e.WokeScore + '</span><p class="desc">' + e.Description + '</p><p class="time">' + e.CreatedAt + '</p></div>' +
      '</div>';
  }

  function setSort(mode) {
    currentSort = mode;
    document.querySelectorAll('.controls button').forEach(b => b.classList.remove('active'));
    document.getElementById('btn-' + mode).classList.add('active');
    sortGrid();
  }

  function sortGrid() {
    const grid = document.getElementById('grid');
    if (!grid) return;
    const cards = Array.from(grid.querySelectorAll('.card'));
    cards.sort((a, b) => {
      if (currentSort === 'score') return Number(b.dataset.score) - Number(a.dataset.score);
      if (currentSort === 'oldest') return a.dataset.time.localeCompare(b.dataset.time);
      return b.dataset.time.localeCompare(a.dataset.time);
    });
    cards.forEach(c => grid.appendChild(c));
  }

  async function poll() {
    try {
      const res = await fetch('/api/entries');
      if (!res.ok) return;
      const entries = await res.json();
      if (!entries || entries.length === 0) return;

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
        '<div><div class="score">' + score + '</div><p style="margin-top:0.3rem">' + desc + '</p><p style="font-size:0.8rem;color:#666;margin-top:0.3rem">' + time + '</p></div>' +
      '</div>' +
      '<div class="modal-rescores"><h3>Re-analyses</h3><div id="rs-list"><p style="color:#666;font-size:0.9rem">Loading...</p></div></div>';

    document.getElementById('modal').classList.add('open');

    try {
      const res = await fetch('/api/entries/' + entryId + '/rescores');
      const rescores = await res.json();
      const list = document.getElementById('rs-list');
      if (!rescores || rescores.length === 0) {
        list.innerHTML = '<p style="color:#666;font-size:0.9rem">No re-analyses yet. Hover the card and click ✨ Rescore.</p>';
        return;
      }
      const card = document.querySelector('.card[data-id="' + entryId + '"]');
      if (card) card.classList.add('has-rescores');
      list.innerHTML = rescores.map(rs =>
        '<div class="rs-card">' +
          '<img src="/photos/' + photo + '" alt="">' +
          '<div><span class="rs-badge">✨ re-analyzed</span><div class="rs-score">' + rs.WokeScore + '</div><div class="rs-subject">' + rs.Subject + '</div><p class="rs-desc">' + rs.Description + '</p></div>' +
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
</script>
</body></html>`))
