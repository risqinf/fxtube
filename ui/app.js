// FX Tube Application Logic - Modular & Hardware-Optimized

let currentTab = 'home';
let currentCategory = '';
let currentSearchQuery = '';
const cache = {};
let currentBrowseVideos = [];
let currentQueue = [];
let currentQueueIndex = -1;
let currentVideo = null;
let currentPlaylistSource = null;
let currentAvailableQualities = ['144', '240', '360', '480', '720', '1080'];

let selectedQuality = localStorage.getItem('fxtube_quality') || '480';
let autoplayEnabled = localStorage.getItem('fxtube_autoplay') !== 'false';
let accountMode = localStorage.getItem('fxtube_account_mode') || 'firefox';
let savedMode = localStorage.getItem('fxtube_player_mode');
if (!savedMode || savedMode === 'mpv') {
  savedMode = 'web';
  localStorage.setItem('fxtube_player_mode', 'web');
}
let playerMode = savedMode;
let streamTimeOffset = 0;
let isTimelineDragging = false;

// Pagination & Infinite Scroll State
let currentPage = 1;
let isFetchingMore = false;
let hasMoreVideos = true;

let mpvPollInterval = null;
let isMpvActive = false;

const player = document.getElementById('embeddedPlayer');

// Clean SVG Icons for reliable cross-platform rendering
const ICONS = {
  play: `<svg viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>`,
  eye: `<svg viewBox="0 0 24 24"><path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/></svg>`,
  sparkles: `<svg viewBox="0 0 24 24"><path d="M12 2l2.4 7.2L21 12l-6.6 2.8L12 22l-2.4-7.2L3 12l6.6-2.8z"/></svg>`,
  flame: `<svg viewBox="0 0 24 24"><path d="M13.5.67s.74 2.65.74 4.8c0 2.06-1.35 3.73-3.41 3.73-2.07 0-3.63-1.67-3.63-3.73l.03-.36C5.21 7.51 4 10.61 4 14c0 4.42 3.58 8 8 8s8-3.58 8-8C20 8.61 17.41 3.8 13.5.67zM11.71 19c-1.78 0-3.22-1.4-3.22-3.14 0-1.62 1.05-2.76 2.81-3.12 1.77-.36 3.6-1.21 4.62-2.58.39 1.29.59 2.65.59 4.04 0 2.65-2.15 4.8-4.8 4.8z"/></svg>`,
  music: `<svg viewBox="0 0 24 24"><path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z"/></svg>`,
  gamepad: `<svg viewBox="0 0 24 24"><path d="M21 6H3c-1.1 0-2 .9-2 2v8c0 1.1.9 2 2 2h18c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2zm-10 7H8v3H6v-3H3v-2h3V8h2v3h3v2zm4.5 2c-.83 0-1.5-.67-1.5-1.5s.67-1.5 1.5-1.5 1.5.67 1.5 1.5-.67 1.5-1.5 1.5zm4-3c-.83 0-1.5-.67-1.5-1.5S18.67 9 19.5 9s1.5.67 1.5 1.5-.67 1.5-1.5 1.5z"/></svg>`,
  news: `<svg viewBox="0 0 24 24"><path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm-5 14H4v-4h11v4zm0-5H4V9h11v4zm5 5h-4V9h4v9z"/></svg>`,
  laptop: `<svg viewBox="0 0 24 24"><path d="M20 18c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2H4c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2H0v2h24v-2h-4zM4 6h16v10H4V6z"/></svg>`,
  mic: `<svg viewBox="0 0 24 24"><path d="M12 14c1.66 0 3-1.34 3-3V5c0-1.66-1.34-3-3-3S9 3.34 9 5v6c0 1.66 1.34 3 3 3zm5.3-3c0 3-2.54 5.1-5.3 5.1S6.7 14 6.7 11H5c0 3.41 2.72 6.23 6 6.72V21h2v-3.28c3.28-.48 6-3.3 6-6.72h-1.7z"/></svg>`
};

// Standalone Mode Detection (runs fully standalone on Android/local without laptop server)
const isStandaloneMode = (window.location.protocol === 'file:') || 
                         (window.location.hostname === '' && !window.location.port) ||
                         (localStorage.getItem('fxtube_standalone') === '1');

function parseDurationToSec(str) {
  if (!str) return 0;
  const parts = str.split(':').map(Number);
  if (parts.length === 3) return (parts[0] || 0) * 3600 + (parts[1] || 0) * 60 + (parts[2] || 0);
  if (parts.length === 2) return (parts[0] || 0) * 60 + (parts[1] || 0);
  return 0;
}

function parseInnerTubeItem(item, isShortsTab = false) {
  if (item.videoWithContextRenderer) {
    const v = item.videoWithContextRenderer;
    const durStr = v.lengthText?.runs?.[0]?.text || '';
    const durSec = parseDurationToSec(durStr);
    const title = v.headline?.runs?.[0]?.text || 'Video';
    const thumbs = v.thumbnail?.thumbnails || [];
    const thumbUrl = thumbs.length > 0 ? thumbs[thumbs.length - 1].url : `https://i.ytimg.com/vi/${v.videoId}/hqdefault.jpg`;
    const channel = v.shortBylineText?.runs?.[0]?.text || v.longBylineText?.runs?.[0]?.text || '';
    const views = v.shortViewCountText?.runs?.[0]?.text || '';
    return {
      id: v.videoId,
      url: `https://www.youtube.com/watch?v=${v.videoId}`,
      title: title,
      uploader: channel,
      thumbnail: thumbUrl,
      duration: durSec,
      duration_str: durStr,
      views: views,
      is_shorts: isShortsTab || title.toLowerCase().includes('#short') || (durSec > 0 && durSec <= 61)
    };
  }
  if (item.videoRenderer) {
    const v = item.videoRenderer;
    const durStr = v.lengthText?.simpleText || '';
    const durSec = parseDurationToSec(durStr);
    const title = v.title?.runs?.[0]?.text || 'Video';
    const thumbs = v.thumbnail?.thumbnails || [];
    const thumbUrl = thumbs.length > 0 ? thumbs[thumbs.length - 1].url : `https://i.ytimg.com/vi/${v.videoId}/hqdefault.jpg`;
    const channel = v.ownerText?.runs?.[0]?.text || '';
    const views = v.shortViewCountText?.simpleText || '';
    return {
      id: v.videoId,
      url: `https://www.youtube.com/watch?v=${v.videoId}`,
      title: title,
      uploader: channel,
      thumbnail: thumbUrl,
      duration: durSec,
      duration_str: durStr,
      views: views,
      is_shorts: isShortsTab || title.toLowerCase().includes('#short') || (durSec > 0 && durSec <= 61)
    };
  }
  return null;
}

async function fetchInnerTubeVideos(query, isShorts = false) {
  const resp = await fetch('https://www.youtube.com/youtubei/v1/search', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      query: query,
      context: {
        client: {
          clientName: 'MWEB',
          clientVersion: '2.20240101.00.00'
        }
      }
    })
  });
  const data = await resp.json();
  const items = data.contents?.sectionListRenderer?.contents?.[0]?.itemSectionRenderer?.contents || [];
  const videos = [];
  for (const it of items) {
    const parsed = parseInnerTubeItem(it, isShorts);
    if (parsed && parsed.id) {
      videos.push(parsed);
    }
  }
  return videos;
}

async function fetchWithTimeout(resource, options = {}) {
  const { timeout = 22000 } = options;
  const controller = new AbortController();
  const id = setTimeout(() => controller.abort(), timeout);
  try {
    const response = await fetch(resource, {
      ...options,
      signal: controller.signal
    });
    return response;
  } finally {
    clearTimeout(id);
  }
}

let toastTimer = null;
function showToast(msg, bg="#2563eb") {
  const t = document.getElementById('toast');
  if (!t) return;
  t.innerText = msg;
  t.style.background = bg;
  t.style.display = 'block';
  if (toastTimer) clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { t.style.display = 'none'; }, 3000);
}

function setTabActive(tabName) {
  document.querySelectorAll('.nav-item').forEach(el => el.classList.remove('active'));
  document.querySelectorAll('.mobile-nav-item').forEach(el => el.classList.remove('active'));

  const mapping = {'home': 0, 'shorts': 1, 'subscriptions': 2, 'history': 3, 'localhistory': 4, 'watchlater': 5, 'liked': 6};
  if (mapping[tabName] !== undefined) {
    const items = document.querySelectorAll('.nav-list .nav-item');
    if (items[mapping[tabName]]) {
      items[mapping[tabName]].classList.add('active');
    }
  }

  const mobileMap = {
    'home': 'mNavHome',
    'shorts': 'mNavShorts',
    'subscriptions': 'mNavSubs',
    'history': 'mNavHistory',
    'localhistory': 'mNavHistory'
  };
  if (mobileMap[tabName]) {
    const mEl = document.getElementById(mobileMap[tabName]);
    if (mEl) mEl.classList.add('active');
  }

  const chipsBar = document.getElementById('homeChipsBar');
  if (chipsBar) {
    chipsBar.style.display = (tabName === 'home') ? 'flex' : 'none';
  }
}

function stopAllShortsPlayers() {
  const shortsContainer = document.getElementById('shortsContainer');
  if (shortsContainer) {
    const iframes = shortsContainer.querySelectorAll('.short-video-frame');
    iframes.forEach(f => { f.src = ''; });
  }
}

function switchToBrowse() {
  const shortsView = document.getElementById('shortsView');
  if (shortsView) shortsView.style.display = 'none';
  stopAllShortsPlayers();

  document.getElementById('watchView').style.display = 'none';
  document.getElementById('browseView').style.display = 'block';
  hidePlayPrompt();
  if (player) {
    player.pause();
    player.removeAttribute('src');
    player.load();
  }
}

function switchToWatch() {
  const shortsView = document.getElementById('shortsView');
  if (shortsView) shortsView.style.display = 'none';
  stopAllShortsPlayers();

  document.getElementById('browseView').style.display = 'none';
  document.getElementById('watchView').style.display = 'block';
}

function switchToShortsView() {
  document.getElementById('browseView').style.display = 'none';
  document.getElementById('watchView').style.display = 'none';
  hidePlayPrompt();
  if (player) {
    try { player.pause(); player.removeAttribute('src'); player.load(); } catch (e) {}
  }
  const shortsView = document.getElementById('shortsView');
  if (shortsView) shortsView.style.display = 'block';
}

let shortsReelList = [];
let shortsPage = 1;
let isFetchingShorts = false;
let shortsObserver = null;

async function loadShortsReel(forceRefresh = false) {
  switchToShortsView();
  setTabActive('shorts');
  currentTab = 'shorts';

  const container = document.getElementById('shortsContainer');
  if (!container) return;

  if (!forceRefresh && shortsReelList.length > 0) {
    activateShortSlide(0);
    return;
  }

  container.innerHTML = `
    <div style="height:100%;display:flex;flex-direction:column;align-items:center;justify-content:center;color:#888;">
      <div class="spinner"></div>
      <p style="margin-top:14px;font-size:14px;color:#aaa;">Menyiapkan Feed Shorts...</p>
    </div>`;

  shortsPage = 1;
  try {
    let videos = [];
    if (isStandaloneMode) {
      videos = await fetchInnerTubeVideos('#shorts trending indonesia', true);
    } else {
      const res = await fetchWithTimeout('/api/feed?tab=shorts&page=1', { timeout: 18000 });
      const data = await res.json();
      videos = data.videos || [];
      if (videos.length === 0) {
        videos = await fetchInnerTubeVideos('#shorts trending indonesia', true);
      }
    }
    shortsReelList = videos;
    renderShortsReel(videos);
  } catch (err) {
    console.warn('Gagal memuat shorts reel via primary:', err);
    try {
      const fallback = await fetchInnerTubeVideos('#shorts indonesia viral', true);
      shortsReelList = fallback;
      renderShortsReel(fallback);
    } catch (e2) {
      container.innerHTML = `
        <div style="height:100%;display:flex;flex-direction:column;align-items:center;justify-content:center;color:#ef4444;padding:20px;text-align:center;">
          <p style="font-size:15px;margin-bottom:12px;">Gagal memuat Shorts: ${escapeHtml(err.message)}</p>
          <button class="header-btn" onclick="loadShortsReel(true)" style="padding:8px 20px;background:#dc2626;color:#fff;">Coba Lagi</button>
        </div>`;
    }
  }
}

function renderShortsReel(videos) {
  const container = document.getElementById('shortsContainer');
  if (!container) return;
  container.innerHTML = '';

  if (!videos || videos.length === 0) {
    container.innerHTML = `
      <div style="height:100%;display:flex;align-items:center;justify-content:center;color:#888;">
        Tidak ada Shorts ditemukan.
      </div>`;
    return;
  }

  videos.forEach((v, index) => {
    const slide = document.createElement('div');
    slide.className = 'short-slide';
    slide.id = `shortSlide_${index}`;
    slide.dataset.index = index;
    slide.dataset.id = v.id;

    slide.innerHTML = `
      <div class="short-card">
        <iframe class="short-video-frame" 
                id="shortFrame_${index}" 
                src="" 
                data-src="https://www.youtube-nocookie.com/embed/${v.id}?autoplay=1&loop=1&playlist=${v.id}&playsinline=1&controls=0&modestbranding=1&rel=0&enablejsapi=1" 
                allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" 
                allowfullscreen>
        </iframe>

        <!-- Gradient Overlay Bottom -->
        <div class="short-overlay-bottom">
          <div class="short-channel-row">
            <img class="short-avatar" src="${v.thumbnail || `https://i.ytimg.com/vi/${v.id}/hqdefault.jpg`}" alt="avatar" loading="lazy">
            <span class="short-channel-name">@${escapeHtml(v.uploader || 'Creator')}</span>
            <button class="short-sub-btn" onclick="toggleShortSub(this)">Subscribe</button>
          </div>
          <div class="short-caption">${escapeHtml(v.title)}</div>
          <div class="short-sound-track">
            <svg viewBox="0 0 24 24"><path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z"/></svg>
            <span class="short-sound-name">${escapeHtml(v.title)} • Audio Asli</span>
          </div>
        </div>

        <!-- Action buttons right side (TikTok style) -->
        <div class="short-action-rail">
          <button class="short-action-btn" onclick="toggleShortLike(this)" title="Suka">
            <div class="short-action-icon">
              <svg viewBox="0 0 24 24"><path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>
            </div>
            <span class="short-action-text">Suka</span>
          </button>

          <button class="short-action-btn" onclick="shareShort('${v.id}')" title="Bagikan">
            <div class="short-action-icon">
              <svg viewBox="0 0 24 24"><path d="M18 16.08c-.76 0-1.44.3-1.96.77L8.91 12.7c.05-.23.09-.46.09-.7s-.04-.47-.09-.7l7.05-4.11c.54.5 1.25.81 2.04.81 1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3c0 .24.04.47.09.7L8.04 9.81C7.5 9.31 6.79 9 6 9c-1.66 0-3 1.34-3 3s1.34 3 3 3c.79 0 1.5-.31 2.04-.81l7.12 4.16c-.05.21-.08.43-.08.65 0 1.61 1.31 2.92 2.92 2.92 1.61 0 2.92-1.31 2.92-2.92s-1.31-2.92-2.92-2.92z"/></svg>
            </div>
            <span class="short-action-text">Bagikan</span>
          </button>

          <button class="short-action-btn" onclick="openShortInWatch('${v.id}')" title="Buka Detail">
            <div class="short-action-icon">
              <svg viewBox="0 0 24 24"><path d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z"/></svg>
            </div>
            <span class="short-action-text">Detail</span>
          </button>

          <button class="short-action-btn short-nav-btn desktop-only" onclick="scrollShortNav(-1)" title="Video Sebelumnya">
            <div class="short-action-icon">
              <svg viewBox="0 0 24 24"><path d="M7.41 15.41L12 10.83l4.59 4.58L18 14l-6-6-6 6z"/></svg>
            </div>
          </button>

          <button class="short-action-btn short-nav-btn desktop-only" onclick="scrollShortNav(1)" title="Video Selanjutnya">
            <div class="short-action-icon">
              <svg viewBox="0 0 24 24"><path d="M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6-6-6 1.41-1.41z"/></svg>
            </div>
          </button>
        </div>
      </div>
    `;
    container.appendChild(slide);
  });

  setupShortsObserver();
  activateShortSlide(0);
}

function setupShortsObserver() {
  if (shortsObserver) shortsObserver.disconnect();
  const container = document.getElementById('shortsContainer');
  if (!container) return;

  shortsObserver = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      const slide = entry.target;
      const idx = parseInt(slide.dataset.index, 10);
      const iframe = slide.querySelector('.short-video-frame');
      if (entry.isIntersecting && entry.intersectionRatio >= 0.6) {
        if (iframe && iframe.src !== iframe.dataset.src) {
          iframe.src = iframe.dataset.src;
        }
        if (idx >= shortsReelList.length - 3 && !isFetchingShorts) {
          loadMoreShorts();
        }
      } else {
        if (iframe && iframe.src) {
          iframe.src = '';
        }
      }
    });
  }, {
    root: container,
    threshold: 0.6
  });

  document.querySelectorAll('.short-slide').forEach(el => shortsObserver.observe(el));
}

function activateShortSlide(index) {
  const slide = document.getElementById(`shortSlide_${index}`);
  if (!slide) return;
  const iframe = slide.querySelector('.short-video-frame');
  if (iframe && iframe.src !== iframe.dataset.src) {
    iframe.src = iframe.dataset.src;
  }
}

async function loadMoreShorts() {
  if (isFetchingShorts) return;
  isFetchingShorts = true;
  shortsPage += 1;
  try {
    const q = `#shorts trending indonesia part ${shortsPage}`;
    const more = await fetchInnerTubeVideos(q, true);
    if (more && more.length > 0) {
      const container = document.getElementById('shortsContainer');
      const startIdx = shortsReelList.length;
      shortsReelList = shortsReelList.concat(more);
      more.forEach((v, i) => {
        const index = startIdx + i;
        const slide = document.createElement('div');
        slide.className = 'short-slide';
        slide.id = `shortSlide_${index}`;
        slide.dataset.index = index;
        slide.dataset.id = v.id;
        slide.innerHTML = `
          <div class="short-card">
            <iframe class="short-video-frame" 
                    id="shortFrame_${index}" 
                    src="" 
                    data-src="https://www.youtube-nocookie.com/embed/${v.id}?autoplay=1&loop=1&playlist=${v.id}&playsinline=1&controls=0&modestbranding=1&rel=0&enablejsapi=1" 
                    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" 
                    allowfullscreen>
            </iframe>
            <div class="short-overlay-bottom">
              <div class="short-channel-row">
                <img class="short-avatar" src="${v.thumbnail || `https://i.ytimg.com/vi/${v.id}/hqdefault.jpg`}" alt="avatar" loading="lazy">
                <span class="short-channel-name">@${escapeHtml(v.uploader || 'Creator')}</span>
                <button class="short-sub-btn" onclick="toggleShortSub(this)">Subscribe</button>
              </div>
              <div class="short-caption">${escapeHtml(v.title)}</div>
              <div class="short-sound-track">
                <svg viewBox="0 0 24 24"><path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z"/></svg>
                <span class="short-sound-name">${escapeHtml(v.title)} • Audio Asli</span>
              </div>
            </div>
            <div class="short-action-rail">
              <button class="short-action-btn" onclick="toggleShortLike(this)" title="Suka">
                <div class="short-action-icon">
                  <svg viewBox="0 0 24 24"><path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>
                </div>
                <span class="short-action-text">Suka</span>
              </button>
              <button class="short-action-btn" onclick="shareShort('${v.id}')" title="Bagikan">
                <div class="short-action-icon">
                  <svg viewBox="0 0 24 24"><path d="M18 16.08c-.76 0-1.44.3-1.96.77L8.91 12.7c.05-.23.09-.46.09-.7s-.04-.47-.09-.7l7.05-4.11c.54.5 1.25.81 2.04.81 1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3c0 .24.04.47.09.7L8.04 9.81C7.5 9.31 6.79 9 6 9c-1.66 0-3 1.34-3 3s1.34 3 3 3c.79 0 1.5-.31 2.04-.81l7.12 4.16c-.05.21-.08.43-.08.65 0 1.61 1.31 2.92 2.92 2.92 1.61 0 2.92-1.31 2.92-2.92s-1.31-2.92-2.92-2.92z"/></svg>
                </div>
                <span class="short-action-text">Bagikan</span>
              </button>
              <button class="short-action-btn" onclick="openShortInWatch('${v.id}')" title="Buka Detail">
                <div class="short-action-icon">
                  <svg viewBox="0 0 24 24"><path d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z"/></svg>
                </div>
                <span class="short-action-text">Detail</span>
              </button>
              <button class="short-action-btn short-nav-btn desktop-only" onclick="scrollShortNav(-1)" title="Video Sebelumnya">
                <div class="short-action-icon">
                  <svg viewBox="0 0 24 24"><path d="M7.41 15.41L12 10.83l4.59 4.58L18 14l-6-6-6 6z"/></svg>
                </div>
              </button>
              <button class="short-action-btn short-nav-btn desktop-only" onclick="scrollShortNav(1)" title="Video Selanjutnya">
                <div class="short-action-icon">
                  <svg viewBox="0 0 24 24"><path d="M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6-6-6 1.41-1.41z"/></svg>
                </div>
              </button>
            </div>
          </div>
        `;
        container.appendChild(slide);
        if (shortsObserver) shortsObserver.observe(slide);
      });
    }
  } catch (err) {
    console.warn('Gagal memuat lebih banyak shorts:', err);
  } finally {
    isFetchingShorts = false;
  }
}

function scrollShortNav(dir) {
  const container = document.getElementById('shortsContainer');
  if (!container) return;
  container.scrollBy({ top: dir * container.clientHeight, behavior: 'smooth' });
}

function toggleShortLike(btn) {
  btn.classList.toggle('liked');
  if (btn.classList.contains('liked')) {
    showToast('Disukai! ❤️', '#ef4444');
  }
}

function toggleShortSub(btn) {
  btn.classList.toggle('subscribed');
  btn.innerText = btn.classList.contains('subscribed') ? 'Disubscribe' : 'Subscribe';
}

function shareShort(videoId) {
  const url = `https://www.youtube.com/shorts/${videoId}`;
  if (navigator.clipboard) {
    navigator.clipboard.writeText(url).then(() => {
      showToast('Link Shorts disalin: ' + url, '#10b981');
    }).catch(() => {
      showToast('Link: ' + url, '#2563eb');
    });
  } else {
    showToast('Link: ' + url, '#2563eb');
  }
}

function openShortInWatch(videoId) {
  const found = shortsReelList.find(s => s.id === videoId);
  const v = found || { id: videoId, title: 'YouTube Short', url: `https://www.youtube.com/shorts/${videoId}`, is_shorts: true };
  currentQueue = shortsReelList;
  currentQueueIndex = shortsReelList.findIndex(s => s.id === videoId);
  playEmbeddedVideo(v);
}

window.addEventListener('keydown', (e) => {
  const shortsView = document.getElementById('shortsView');
  if (shortsView && shortsView.style.display !== 'none') {
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      scrollShortNav(-1);
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      scrollShortNav(1);
    }
  }
});

function switchToWatchViewDirect() {
  if (currentQueue && currentQueueIndex >= 0 && currentQueueIndex < currentQueue.length) {
    stopMPV();
    playDirectEmbedded(currentQueueIndex, false);
  } else {
    switchToWatch();
  }
}

function showLoadMore(show) {
  const box = document.getElementById('loadMoreBox');
  if (box) box.style.display = show ? 'block' : 'none';
}

function hideLoadMore() {
  showLoadMore(false);
}

// --- TAB LOADING ---
async function loadTab(tabName, forceRefresh=false) {
  if (tabName === 'shorts') {
    loadShortsReel(forceRefresh);
    return;
  }

  switchToBrowse();
  currentTab = tabName;
  currentPage = 1;
  hasMoreVideos = true;
  hideLoadMore();

  if (tabName === 'home' && !forceRefresh) {
    currentCategory = '';
    document.querySelectorAll('.chips-bar .chip').forEach((c, i) => {
      c.classList.toggle('active', i === 0);
    });
  }
  setTabActive(tabName);

  const titles = {
    home: 'Home (Beranda)',
    shorts: 'YouTube Shorts (Video Pendek)',
    subscriptions: 'Subscriptions (Feed Akun Anda)',
    history: 'Riwayat Tontonan (Akun YouTube)',
    localhistory: 'Riwayat Putar Lokal (FX Tube)',
    watchlater: 'Watch Later (Tonton Nanti)',
    liked: 'Liked Videos (Video Disukai)',
  };
  document.getElementById('sectionTitle').innerText = titles[tabName] || tabName;

  if (tabName === 'localhistory') {
    const localItems = JSON.parse(localStorage.getItem('fxtube_local_history') || '[]');
    cache[tabName] = localItems;
    renderVideos(localItems);
    return;
  }

  if (accountMode === 'guest' && ['subscriptions', 'history', 'watchlater', 'liked'].includes(tabName)) {
    document.getElementById('videoGrid').innerHTML = `
      <div style="text-align:center;padding:60px 20px;grid-column:1/-1;">
        <svg viewBox="0 0 24 24" style="width:48px;height:48px;fill:#888;margin-bottom:12px;">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z"/>
        </svg>
        <h3 style="color:#eee;font-size:16px;margin-bottom:8px;">Tab Ini Membutuhkan Akun YouTube</h3>
        <p style="color:#888;font-size:13px;max-width:440px;margin:0 auto 16px;line-height:1.5;">
          Anda sedang dalam <strong>Mode Tamu (Publik)</strong>. Untuk melihat ${titles[tabName]}, silakan buka Pengaturan dan pilih <strong>Sinkronkan ke Firefox</strong>.
        </p>
        <button class="btn-card-mpv" onclick="openSettingsModal()" style="display:inline-flex;padding:9px 22px;font-size:13px;background:#2563eb;">
          ${ICONS.play} Buka Pengaturan Akun
        </button>
      </div>`;
    return;
  }

  const cacheKey = tabName + '::' + currentCategory;
  if (!forceRefresh && cache[cacheKey]) {
    renderVideos(cache[cacheKey]);
    if (cache[cacheKey].length >= 10) showLoadMore(true);
    return;
  }

  showLoader(true, `Mengambil ${titles[tabName]}...`);
  try {
    if (isStandaloneMode) {
      let query = 'trending indonesia';
      let isShorts = false;
      if (tabName === 'shorts') {
        query = '#shorts trending indonesia';
        isShorts = true;
      } else if (tabName === 'home') {
        if (currentCategory && currentCategory !== 'Semua') {
          query = currentCategory + ' indonesia';
        }
      } else if (tabName === 'history' || tabName === 'localhistory') {
        const hist = getLocalHistory();
        cache[cacheKey] = hist;
        renderVideos(hist);
        return;
      } else if (tabName === 'subscriptions') {
        query = 'indonesia trending viral';
      }
      const videos = await fetchInnerTubeVideos(query, isShorts);
      cache[cacheKey] = videos;
      renderVideos(videos);
      if (videos.length >= 10) showLoadMore(true);
      return;
    }

    let url = `/api/feed?tab=${tabName}&page=1`;
    if (forceRefresh) url += '&refresh=1';
    if (tabName === 'home' && currentCategory) {
      url += `&category=${encodeURIComponent(currentCategory)}`;
    }
    const res = await fetchWithTimeout(url, { timeout: 22000 });
    const data = await res.json();
    if (data.error) throw new Error(data.error);
    cache[cacheKey] = data.videos || [];
    renderVideos(cache[cacheKey]);
    if (cache[cacheKey].length >= 10) {
      showLoadMore(true);
    }
  } catch (err) {
    const msg = err.name === 'AbortError' ? 'Koneksi timeout (jaringan atau yt-dlp lambat).' : err.message;
    document.getElementById('videoGrid').innerHTML = `
      <div style="text-align:center;padding:50px 20px;grid-column:1/-1;">
        <p style="color:#ef4444;font-size:15px;margin-bottom:14px;">Gagal memuat: ${escapeHtml(msg)}</p>
        <button class="btn-card-mpv" onclick="refreshCurrentTab()" style="display:inline-flex;padding:8px 20px;font-size:13px;background:#2563eb;">
          ${ICONS.play} Coba Lagi
        </button>
      </div>`;
  } finally {
    showLoader(false);
  }
}

// --- PAGINATION & INFINITE SCROLL ---
async function loadMoreVideos() {
  if (isFetchingMore || !hasMoreVideos) return;
  if (currentTab === 'localhistory') return;

  isFetchingMore = true;
  const btn = document.getElementById('btnLoadMore');
  if (btn) btn.innerText = 'Memuat video lebih banyak...';

  const nextPage = currentPage + 1;
  try {
    if (isStandaloneMode) {
      let q = currentTab === 'search' ? currentSearchQuery : (currentTab === 'shorts' ? '#shorts trending indonesia' : (currentCategory && currentCategory !== 'Semua' ? currentCategory + ' indonesia' : 'trending indonesia'));
      q += ` bagian ${nextPage}`;
      const moreVideos = await fetchInnerTubeVideos(q, currentTab === 'shorts');
      if (moreVideos && moreVideos.length > 0) {
        currentPage = nextPage;
        appendVideos(moreVideos);
      } else {
        hasMoreVideos = false;
        if (btn) btn.innerText = 'Semua video telah ditampilkan';
      }
      return;
    }
    let url = '';
    if (currentTab === 'search') {
      url = `/api/search?q=${encodeURIComponent(currentSearchQuery)}&page=${nextPage}`;
    } else {
      url = `/api/feed?tab=${currentTab}&page=${nextPage}`;
      if (currentTab === 'home' && currentCategory) {
        url += `&category=${encodeURIComponent(currentCategory)}`;
      }
    }

    const res = await fetchWithTimeout(url, { timeout: 22000 });
    const data = await res.json();
    if (data.videos && data.videos.length > 0) {
      currentPage = nextPage;
      appendVideos(data.videos);
    } else {
      hasMoreVideos = false;
      if (btn) btn.innerText = 'Semua video telah ditampilkan';
    }
  } catch (err) {
    console.warn('Gagal memuat halaman berikutnya:', err);
    if (btn) btn.innerText = 'Coba lagi memuat video';
  } finally {
    isFetchingMore = false;
    if (hasMoreVideos && btn) {
      btn.innerHTML = `<svg viewBox="0 0 24 24" style="width:16px;height:16px;fill:currentColor;margin-right:6px;"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg> Muat Lebih Banyak Video`;
    }
  }
}

function appendVideos(newVideos) {
  const grid = document.getElementById('videoGrid');
  if (!grid || !newVideos || newVideos.length === 0) return;

  const startIndex = currentBrowseVideos.length;
  currentBrowseVideos = currentBrowseVideos.concat(newVideos);

  newVideos.forEach((v, i) => {
    const cardIndex = startIndex + i;
    const card = document.createElement('div');
    card.className = 'video-card';

    let thumb = 'https://i.ytimg.com/img/no_thumbnail.jpg';
    if (v.thumbnails && v.thumbnails.length > 0) {
      thumb = v.thumbnails[v.thumbnails.length - 1].url || thumb;
    }
    const durationText = v.duration_string || formatDuration(v.duration);
    const isShort = v.is_shorts || (v.title && (v.title.toLowerCase().includes('#shorts') || v.title.toLowerCase().includes('#short'))) || (v.duration > 0 && v.duration <= 61);

    card.innerHTML = `
      <div class="thumb-wrapper" onclick="handleCardClick(${cardIndex})">
        <img class="thumb-img" src="${thumb}" loading="lazy" decoding="async" alt="thumb">
        ${isShort ? `<span class="shorts-badge">${ICONS.flame} Shorts</span>` : ''}
        ${durationText ? `<span class="duration-badge">${durationText}</span>` : ''}
      </div>
      <div class="card-info" onclick="handleCardClick(${cardIndex})">
        <div class="card-title" title="${escapeHtml(v.title)}">${escapeHtml(v.title)}</div>
        <div class="card-channel">${escapeHtml(v.channel || v.uploader || 'YouTube')}</div>
        <div class="card-footer">
          <button class="btn-card-web" onclick="event.stopPropagation(); playDirectEmbedded(${cardIndex}, true)" title="Putar di Web Browser (Default)">
            ${ICONS.eye} Putar Web
          </button>
          <button class="btn-card-mpv" onclick="event.stopPropagation(); playDirectMPV(${cardIndex}, true)" title="Putar di MPV Native (~40MB RAM)">
            ${ICONS.play} MPV
          </button>
        </div>
      </div>
    `;
    grid.appendChild(card);
  });
}

function filterCategory(cat, btn) {
  document.querySelectorAll('.chips-bar .chip').forEach(c => c.classList.remove('active'));
  btn.classList.add('active');
  currentCategory = cat;
  loadTab('home', true);
}

function refreshCurrentTab() {
  if (currentTab === 'search') {
    executeSearch(true);
  } else {
    loadTab(currentTab, true);
  }
}

async function executeSearch(forceRefresh=false) {
  const query = document.getElementById('searchInput').value.trim();
  if (!query) return;

  currentSearchQuery = query;
  currentPage = 1;
  hasMoreVideos = true;
  hideLoadMore();

  switchToBrowse();
  currentTab = 'search';
  document.querySelectorAll('.nav-item').forEach(el => el.classList.remove('active'));
  document.getElementById('homeChipsBar').style.display = 'none';
  document.getElementById('sectionTitle').innerText = `Hasil Pencarian: "${query}"`;

  showLoader(true, `Mencari "${query}"...`);
  try {
    if (isStandaloneMode) {
      const isShorts = query.toLowerCase().includes('#short');
      const videos = await fetchInnerTubeVideos(query, isShorts);
      currentQueue = videos;
      renderVideos(videos);
      if (videos.length >= 10) showLoadMore(true);
      return;
    }

    let url = `/api/search?q=${encodeURIComponent(query)}&page=1`;
    if (forceRefresh) url += '&refresh=1';
    const res = await fetchWithTimeout(url, { timeout: 22000 });
    const data = await res.json();
    if (data.error) throw new Error(data.error);
    renderVideos(data.videos || []);
    if (data.videos && data.videos.length >= 10) {
      showLoadMore(true);
    }
  } catch (err) {
    const msg = err.name === 'AbortError' ? 'Pencarian timeout.' : err.message;
    document.getElementById('videoGrid').innerHTML = `
      <div style="text-align:center;padding:50px 20px;grid-column:1/-1;">
        <p style="color:#ef4444;font-size:15px;margin-bottom:14px;">Pencarian gagal: ${escapeHtml(msg)}</p>
        <button class="btn-card-mpv" onclick="executeSearch(true)" style="display:inline-flex;padding:8px 20px;font-size:13px;background:#2563eb;">
          ${ICONS.play} Coba Lagi
        </button>
      </div>`;
  } finally {
    showLoader(false);
  }
}

function showLoader(show, text="Memuat...") {
  document.getElementById('loader').style.display = show ? 'flex' : 'none';
  document.getElementById('videoGrid').style.display = show ? 'none' : 'grid';
  document.getElementById('loaderText').innerText = text;
}

function formatDuration(sec) {
  if (!sec) return '';
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  const s = sec % 60;
  if (h > 0) return `${h}:${m < 10 ? '0' : ''}${m}:${s < 10 ? '0' : ''}${s}`;
  return `${m}:${s < 10 ? '0' : ''}${s}`;
}

function renderVideos(videos) {
  const grid = document.getElementById('videoGrid');
  grid.innerHTML = '';
  if (!videos || videos.length === 0) {
    grid.innerHTML = '<p style="color:#888;grid-column:1/-1;padding:40px 0;text-align:center;">Tidak ada video ditemukan.</p>';
    hideLoadMore();
    return;
  }

  currentBrowseVideos = videos;

  videos.forEach((v, index) => {
    const card = document.createElement('div');
    card.className = 'video-card';

    let thumb = 'https://i.ytimg.com/img/no_thumbnail.jpg';
    if (v.thumbnails && v.thumbnails.length > 0) {
      thumb = v.thumbnails[v.thumbnails.length - 1].url || thumb;
    }

    const durationText = v.duration_string || formatDuration(v.duration);
    const isShort = v.is_shorts || (v.title && (v.title.toLowerCase().includes('#shorts') || v.title.toLowerCase().includes('#short'))) || (v.duration > 0 && v.duration <= 61);

    card.innerHTML = `
      <div class="thumb-wrapper" onclick="handleCardClick(${index})">
        <img class="thumb-img" src="${thumb}" loading="lazy" decoding="async" alt="thumb">
        ${isShort ? `<span class="shorts-badge">${ICONS.flame} Shorts</span>` : ''}
        ${durationText ? `<span class="duration-badge">${durationText}</span>` : ''}
      </div>
      <div class="card-info" onclick="handleCardClick(${index})">
        <div class="card-title" title="${escapeHtml(v.title)}">${escapeHtml(v.title)}</div>
        <div class="card-channel">${escapeHtml(v.channel || v.uploader || 'YouTube')}</div>
        <div class="card-footer">
          <button class="btn-card-web" onclick="event.stopPropagation(); playDirectEmbedded(${index}, true)" title="Putar di Web Browser (Default)">
            ${ICONS.eye} Putar Web
          </button>
          <button class="btn-card-mpv" onclick="event.stopPropagation(); playDirectMPV(${index}, true)" title="Putar di MPV Native (~40MB RAM)">
            ${ICONS.play} MPV
          </button>
        </div>
      </div>
    `;
    grid.appendChild(card);
  });
}

function escapeHtml(str) {
  if (!str) return '';
  return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;');
}

// --- PLAYBACK ROUTER ---
function handleCardClick(index) {
  if (playerMode === 'mpv') {
    playDirectMPV(index, true);
  } else {
    playDirectEmbedded(index, true);
  }
}

function recordLocalHistory(v) {
  try {
    let history = JSON.parse(localStorage.getItem('fxtube_local_history') || '[]');
    history = history.filter(item => item.url !== v.url);
    history.unshift({
      id: v.id,
      title: v.title,
      url: v.url,
      channel: v.channel || v.uploader,
      duration: v.duration,
      duration_string: v.duration_string,
      thumbnails: v.thumbnails,
      watchedAt: new Date().toISOString()
    });
    if (history.length > 50) history = history.slice(0, 50);
    localStorage.setItem('fxtube_local_history', JSON.stringify(history));
  } catch (e) {}
}

// --- YOUTUBE MIX / PLAYLIST LOADER ---
async function loadMixPlaylistForVideo(v) {
  if (!v) return;
  const sourceKey = (v.url && v.url.includes('list=')) ? v.url : (v.id || v.url);
  if (currentPlaylistSource === sourceKey && currentQueue.length > 1) {
    const idx = currentQueue.findIndex(item => (v.id && item.id === v.id) || (v.url && item.url === v.url));
    if (idx !== -1) currentQueueIndex = idx;
    renderQueueList();
    return;
  }

  currentPlaylistSource = sourceKey;
  currentQueue = [v];
  currentQueueIndex = 0;
  const subtitle = document.getElementById('queueSubtitle');
  if (subtitle) subtitle.innerText = 'Mengambil daftar putar / Mix...';
  renderQueueList();

  try {
    const res = await fetchWithTimeout(`/api/playlist?id=${encodeURIComponent(v.id || '')}&url=${encodeURIComponent(v.url || '')}`, { timeout: 16000 });
    const data = await res.json();
    if (data && data.videos && data.videos.length > 0) {
      let idx = data.videos.findIndex(item => (v.id && item.id === v.id) || (v.url && item.url === v.url));
      if (idx === -1) {
        data.videos.unshift(v);
        idx = 0;
      }
      currentQueue = data.videos;
      currentQueueIndex = idx;
      const sub = document.getElementById('queueSubtitle');
      if (sub) {
        const isRealPlaylist = v.url && v.url.includes('list=') && !v.url.includes('list=RD');
        sub.innerText = (isRealPlaylist ? 'Playlist' : 'YouTube Mix') + ` (${currentQueue.length} video)`;
      }
      renderQueueList();

      // Preload next track immediately
      if (idx + 1 < currentQueue.length) {
        const nextV = currentQueue[idx + 1];
        fetch(`/api/stream/preload?url=${encodeURIComponent(nextV.url)}&quality=${selectedQuality}`);
      }
    } else {
      const sub = document.getElementById('queueSubtitle');
      if (sub) sub.innerText = '1 video';
    }
  } catch (err) {
    console.warn('Gagal memuat mix:', err);
    const sub = document.getElementById('queueSubtitle');
    if (sub) sub.innerText = '1 video (Mix offline)';
  }
}

// --- MPV PLAYBACK ---
async function playDirectMPV(index, fromBrowse = false) {
  let v = null;
  if (fromBrowse) {
    if (!currentBrowseVideos || index < 0 || index >= currentBrowseVideos.length) return;
    v = currentBrowseVideos[index];
    loadMixPlaylistForVideo(v);
  } else {
    if (!currentQueue || index < 0 || index >= currentQueue.length) return;
    v = currentQueue[index];
    currentQueueIndex = index;
    renderQueueList();
  }

  if (!v) return;
  currentVideo = v;
  recordLocalHistory(v);

  hidePlayPrompt();
  if (player) {
    player.pause();
    player.removeAttribute('src');
    player.load();
  }
  showToast(`Membuka di MPV (${selectedQuality}p)...`);
  showMpvBar(v);

  try {
    const res = await fetch('/api/play', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        url: v.url,
        title: v.title,
        mode: 'video',
        quality: selectedQuality
      })
    });
    const data = await res.json();
    if (data.status === 'error') {
      showToast('Gagal memanggil MPV: ' + (data.error || ''), '#dc2626');
      hideMpvBar();
    } else {
      startMpvPolling();
    }
  } catch (e) {
    showToast('Gagal terhubung ke backend MPV', '#dc2626');
    hideMpvBar();
  }
}

function showMpvBar(v) {
  const bar = document.getElementById('mpvControllerBar');
  document.getElementById('mpvBarTitle').innerText = v.title;
  document.getElementById('mpvBarChannel').innerText = v.channel || v.uploader || 'YouTube';
  document.getElementById('mpvBtnPause').innerHTML = `<svg viewBox="0 0 24 24" style="width:14px;height:14px;fill:currentColor"><path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/></svg>`;
  bar.style.display = 'flex';
  document.getElementById('contentArea').classList.add('has-mpv-bar');
  isMpvActive = true;
}

function hideMpvBar() {
  document.getElementById('mpvControllerBar').style.display = 'none';
  document.getElementById('contentArea').classList.remove('has-mpv-bar');
  isMpvActive = false;
  stopMpvPolling();
}

function startMpvPolling() {
  stopMpvPolling();
  mpvPollInterval = setInterval(async () => {
    try {
      const res = await fetch('/api/player/status');
      const data = await res.json();
      if (data && data.playing) {
        isMpvActive = true;
        const btn = document.getElementById('mpvBtnPause');
        if (data.paused) {
          btn.innerHTML = `<svg viewBox="0 0 24 24" style="width:14px;height:14px;fill:currentColor"><path d="M8 5v14l11-7z"/></svg>`;
        } else {
          btn.innerHTML = `<svg viewBox="0 0 24 24" style="width:14px;height:14px;fill:currentColor"><path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/></svg>`;
        }
      } else {
        if (isMpvActive) {
          isMpvActive = false;
          stopMpvPolling();
          hideMpvBar();
          if (autoplayEnabled && currentQueueIndex + 1 < currentQueue.length) {
            showToast('Video selesai, memutar antrean berikutnya...');
            playDirectMPV(currentQueueIndex + 1);
          }
        }
      }
    } catch (e) {}
  }, 1500);
}

function stopMpvPolling() {
  if (mpvPollInterval) {
    clearInterval(mpvPollInterval);
    mpvPollInterval = null;
  }
}

async function toggleMPVPause() {
  try {
    await fetchWithTimeout('/api/player/command', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({ command: 'toggle_pause' }),
      timeout: 2500
    });
  } catch (e) {}
}

async function seekMPV(delta) {
  try {
    await fetchWithTimeout('/api/player/command', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({ command: delta > 0 ? 'seek_forward' : 'seek_backward' }),
      timeout: 2500
    });
    showToast(delta > 0 ? '+10 detik' : '-10 detik');
  } catch (e) {}
}

async function stopMPV() {
  try {
    await fetchWithTimeout('/api/player/command', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({ command: 'stop' }),
      timeout: 2500
    });
  } catch (e) {}
  hideMpvBar();
  showToast('Pemutaran MPV dihentikan');
}

// --- DYNAMIC QUALITY PICKER (144p - 1080p) ---
function renderQualityButtons(qualities) {
  const container = document.getElementById('qualityPicker');
  if (!container) return;
  container.innerHTML = '';

  const list = (qualities && qualities.length > 0) ? qualities : ['144', '240', '360', '480', '720', '1080'];
  currentAvailableQualities = list;

  // If selected quality is not in available qualities, select closest lower or available
  if (!list.includes(selectedQuality)) {
    const selNum = parseInt(selectedQuality, 10);
    const valid = list.map(q => parseInt(q, 10)).filter(q => q <= selNum);
    if (valid.length > 0) {
      selectedQuality = String(Math.max(...valid));
    } else {
      selectedQuality = list[0];
    }
  }

  list.forEach(q => {
    const btn = document.createElement('button');
    btn.className = 'btn-quality' + (selectedQuality === q ? ' active' : '');
    btn.title = `${q}p H.264 (Maksimal 1080p)`;
    btn.innerHTML = `${ICONS.play} ${q}p`;
    btn.onclick = () => switchQuality(q);
    container.appendChild(btn);
  });

  const badge = document.getElementById('watchQualityBadge');
  if (badge) badge.innerText = `${selectedQuality}p H.264 • Low CPU`;
  const opt = document.getElementById('specOptText');
  if (opt) opt.innerText = `${selectedQuality}p H.264`;
}

// --- EMBEDDED WEB PLAYER ---
function playDirectEmbedded(index, fromBrowse = false) {
  let v = null;
  if (fromBrowse) {
    if (!currentBrowseVideos || index < 0 || index >= currentBrowseVideos.length) return;
    v = currentBrowseVideos[index];
    loadMixPlaylistForVideo(v);
  } else {
    if (!currentQueue || index < 0 || index >= currentQueue.length) return;
    v = currentQueue[index];
    currentQueueIndex = index;
    renderQueueList();
  }

  if (!v) return;
  playEmbeddedVideo(v);
}

function showPlayPrompt() {
  const el = document.getElementById('playPromptOverlay');
  if (el) el.style.display = 'flex';
}

function hidePlayPrompt() {
  const el = document.getElementById('playPromptOverlay');
  if (el) el.style.display = 'none';
}

function userResumePlayback() {
  hidePlayPrompt();
  if (player) {
    player.play().catch(e => console.warn('User resume play:', e));
  }
}

function playViaEmbedFallback(v) {
  const standaloneIframe = document.getElementById('standalonePlayer');
  if (!standaloneIframe) return;
  if (player) {
    try { player.pause(); player.removeAttribute('src'); player.load(); } catch (e) {}
    player.style.display = 'none';
  }
  standaloneIframe.style.display = 'block';
  standaloneIframe.src = `https://www.youtube-nocookie.com/embed/${v.id}?autoplay=1&playsinline=1&rel=0&modestbranding=1`;
  const timelineBox = document.querySelector('.web-timeline-box');
  if (timelineBox) timelineBox.style.display = 'none';
  const qualityBox = document.getElementById('watchQualityBadge');
  if (qualityBox) qualityBox.innerText = 'YouTube Embed (Cadangan Otomatis)';
  showToast('Memutar via YouTube Embed (Stream cadangan)', '#2563eb');
}

async function playEmbeddedVideo(v) {
  currentVideo = v;
  streamTimeOffset = 0;
  switchToWatch();
  recordLocalHistory(v);
  hideMpvBar();
  hidePlayPrompt();

  document.getElementById('watchTitle').innerText = v.title;
  document.getElementById('watchChannel').innerText = v.channel || v.uploader || 'YouTube';

  const isShort = v.is_shorts || (v.title && (v.title.toLowerCase().includes('#shorts') || v.title.toLowerCase().includes('#short'))) || (v.duration > 0 && v.duration <= 61);
  const container = document.querySelector('.video-container');
  if (container) {
    if (isShort) {
      container.classList.add('shorts-mode');
      if (player) player.loop = true;
    } else {
      container.classList.remove('shorts-mode');
      if (player) player.loop = false;
    }
  }

  // 1. Detect and render available resolutions (up to 1080p)
  renderQualityButtons(currentAvailableQualities);
  fetch(`/api/video/info?url=${encodeURIComponent(v.url)}`)
    .then(r => r.json())
    .then(data => {
      if (data && data.qualities && data.qualities.length > 0) {
        renderQualityButtons(data.qualities);
      }
    })
    .catch(() => {});

  const totalSec = v.duration || 0;
  const curEl = document.getElementById('webTimeCurrent');
  const totEl = document.getElementById('webTimeTotal');
  const timeline = document.getElementById('webTimeline');
  if (curEl) curEl.innerText = '0:00';
  if (totEl) totEl.innerText = formatDuration(totalSec);
  if (timeline) {
    timeline.max = totalSec > 0 ? totalSec : 100;
    timeline.value = 0;
  }

  const overlay = document.getElementById('playerLoader');
  const overlayText = document.getElementById('playerLoaderText');
  if (overlay) overlay.style.display = 'flex';
  if (overlayText) overlayText.innerText = `Menyiapkan stream ${selectedQuality}p H.264...`;

  let loaded = false;
  const hideOverlay = () => {
    if (!loaded) {
      loaded = true;
      if (overlay) overlay.style.display = 'none';
    }
  };

  const timer = setTimeout(hideOverlay, 8000);

  // Cleanly reset previous player state to prevent browser abort error cascades
  try {
    player.pause();
    player.removeAttribute('src');
    player.load();
  } catch (e) {}

  player.onloadedmetadata = () => {
    clearTimeout(timer);
    hideOverlay();
  };
  player.oncanplay = () => {
    clearTimeout(timer);
    hideOverlay();
  };
  player.onerror = () => {
    clearTimeout(timer);
    hideOverlay();
    if (!player.src || player.src === '' || player.src.endsWith('/')) return;
    console.warn('HTML5 stream playback error, seamlessly falling back to clean embed player for', v.id);
    playViaEmbedFallback(v);
  };

  player.ontimeupdate = () => {
    const currentActual = streamTimeOffset + (player.currentTime || 0);
    const total = currentVideo ? (currentVideo.duration || 0) : 0;
    const curSpan = document.getElementById('webTimeCurrent');
    const totSpan = document.getElementById('webTimeTotal');
    const tSlider = document.getElementById('webTimeline');
    if (curSpan) curSpan.innerText = formatDuration(Math.floor(currentActual));
    if (totSpan && total > 0) totSpan.innerText = formatDuration(Math.floor(total));
    if (tSlider && total > 0 && !isTimelineDragging) {
      tSlider.max = total;
      tSlider.value = Math.floor(currentActual);
    }
  };

  // Pre-resolve next video in background to eliminate next-video wait times
  if (currentQueue && currentQueueIndex + 1 < currentQueue.length) {
    const nextV = currentQueue[currentQueueIndex + 1];
    fetch(`/api/stream/preload?url=${encodeURIComponent(nextV.url)}&quality=${selectedQuality}`);
  }

  const standaloneIframe = document.getElementById('standalonePlayer');
  if (isStandaloneMode && standaloneIframe) {
    clearTimeout(timer);
    hideOverlay();
    if (player) {
      try { player.pause(); player.removeAttribute('src'); player.load(); } catch (e) {}
      player.style.display = 'none';
    }
    standaloneIframe.style.display = 'block';
    standaloneIframe.src = `https://www.youtube-nocookie.com/embed/${v.id}?autoplay=1&playsinline=1&rel=0&modestbranding=1`;
    const timelineBox = document.querySelector('.web-timeline-box');
    if (timelineBox) timelineBox.style.display = 'none';
    const qualityBox = document.getElementById('watchQualityBadge');
    if (qualityBox) qualityBox.innerText = 'YouTube Embed • Standalone';
    return;
  } else {
    if (standaloneIframe) {
      standaloneIframe.style.display = 'none';
      standaloneIframe.src = '';
    }
    if (player) player.style.display = 'block';
    const timelineBox = document.querySelector('.web-timeline-box');
    if (timelineBox) timelineBox.style.display = 'flex';
  }

  try {
    const streamUrl = `/api/stream?url=${encodeURIComponent(v.url)}&quality=${selectedQuality}&_t=${Date.now()}`;
    player.src = streamUrl;
    player.load();
    const playPromise = player.play();
    if (playPromise !== undefined) {
      playPromise.catch(e => {
        console.log('Autoplay deferred or blocked by browser policy:', e);
        showPlayPrompt();
      });
    }
  } catch (err) {
    clearTimeout(timer);
    hideOverlay();
    showToast('Gagal memuat video: ' + err.message, '#dc2626');
  }
}

function seekWeb(delta) {
  if (!currentVideo) return;
  const total = currentVideo.duration || 0;
  const currentActual = streamTimeOffset + (player ? (player.currentTime || 0) : 0);
  let target = Math.max(0, currentActual + delta);
  if (total > 0 && target >= total) target = Math.max(0, total - 2);
  seekWebTo(Math.floor(target));
}

function seekWebTo(seconds) {
  if (!currentVideo) return;
  streamTimeOffset = seconds;
  const curEl = document.getElementById('webTimeCurrent');
  if (curEl) curEl.innerText = formatDuration(seconds);
  const timeline = document.getElementById('webTimeline');
  if (timeline) timeline.value = seconds;

  showToast(`Melompat ke ${formatDuration(seconds)}`);

  const overlay = document.getElementById('playerLoader');
  const overlayText = document.getElementById('playerLoaderText');
  if (overlay) overlay.style.display = 'flex';
  if (overlayText) overlayText.innerText = `Melompat ke ${formatDuration(seconds)}...`;

  let seekDone = false;
  const hideOverlay = () => {
    if (!seekDone) {
      seekDone = true;
      if (overlay) overlay.style.display = 'none';
    }
  };

  const timer = setTimeout(hideOverlay, 6000);
  const onReady = () => {
    clearTimeout(timer);
    hideOverlay();
    player.play().catch(() => {});
    player.removeEventListener('canplay', onReady);
    player.removeEventListener('loadedmetadata', onReady);
  };
  player.addEventListener('canplay', onReady);
  player.addEventListener('loadedmetadata', onReady);

  player.src = `/api/stream?url=${encodeURIComponent(currentVideo.url)}&quality=${selectedQuality}&start=${seconds}&_t=${Date.now()}`;
  player.load();
  player.play().catch(() => {});
}

function onWebTimelineInput(val) {
  isTimelineDragging = true;
  const curEl = document.getElementById('webTimeCurrent');
  if (curEl) curEl.innerText = formatDuration(parseInt(val, 10));
}

function onWebTimelineChange(val) {
  isTimelineDragging = false;
  seekWebTo(parseInt(val, 10));
}

function switchQuality(q) {
  if (selectedQuality === q) return;
  selectedQuality = q;
  localStorage.setItem('fxtube_quality', q);
  renderQualityButtons(currentAvailableQualities);
  if (!currentVideo) return;

  if (isMpvActive) {
    playDirectMPV(currentQueueIndex >= 0 ? currentQueueIndex : 0, false);
    return;
  }

  const currentActual = streamTimeOffset + (player.currentTime || 0);
  const wasPlaying = !player.paused;

  const overlay = document.getElementById('playerLoader');
  const overlayText = document.getElementById('playerLoaderText');
  if (overlay) overlay.style.display = 'flex';
  if (overlayText) overlayText.innerText = `Beralih ke kualitas ${q}p...`;

  let qDone = false;
  const hideOverlay = () => {
    if (!qDone) {
      qDone = true;
      if (overlay) overlay.style.display = 'none';
    }
  };

  const timer = setTimeout(hideOverlay, 7000);

  const onReady = () => {
    clearTimeout(timer);
    hideOverlay();
    if (wasPlaying) player.play().catch(() => {});
    player.removeEventListener('loadedmetadata', onReady);
    player.removeEventListener('canplay', onReady);
  };

  player.addEventListener('loadedmetadata', onReady);
  player.addEventListener('canplay', onReady);

  player.onerror = () => {
    clearTimeout(timer);
    hideOverlay();
    showToast(`Kualitas ${q}p gagal di browser.`, '#ef4444');
  };

  const startOffset = Math.floor(currentActual);
  streamTimeOffset = startOffset;
  const streamUrl = `/api/stream?url=${encodeURIComponent(currentVideo.url)}&quality=${q}&start=${startOffset}&_t=${Date.now()}`;
  player.src = streamUrl;
  player.load();
  player.play().catch(() => {});

  showToast(`Kualitas diubah ke ${q}p`);
}

function updateQualityButtonsUI() {
  renderQualityButtons(currentAvailableQualities);

  const modeLabels = {
    'firefox': 'Firefox Sync',
    'guest': 'Mode Tamu',
    'custom': 'cookies.txt'
  };
  const accBadge = document.getElementById('accountModeBadge');
  if (accBadge) accBadge.innerText = modeLabels[accountMode] || 'Active';

  const playerLabels = {
    'mpv': 'MPV Native (~40MB)',
    'web': 'Embedded Web'
  };
  const playBadge = document.getElementById('playerModeBadge');
  if (playBadge) playBadge.innerText = playerLabels[playerMode] || 'MPV Native';
}

function playNextInQueue() {
  if (!currentQueue || currentQueue.length === 0) return;
  if (currentQueueIndex + 1 < currentQueue.length) {
    if (isMpvActive || playerMode === 'mpv') {
      playDirectMPV(currentQueueIndex + 1, false);
    } else {
      playDirectEmbedded(currentQueueIndex + 1, false);
    }
  } else {
    showToast('Sudah di video terakhir dalam antrean.');
  }
}

function playPrevInQueue() {
  if (!currentQueue || currentQueue.length === 0) return;
  if (currentQueueIndex > 0) {
    if (isMpvActive || playerMode === 'mpv') {
      playDirectMPV(currentQueueIndex - 1, false);
    } else {
      playDirectEmbedded(currentQueueIndex - 1, false);
    }
  } else {
    showToast('Sudah di video pertama dalam antrean.');
  }
}

player.onended = () => {
  if (autoplayEnabled && currentQueueIndex + 1 < currentQueue.length) {
    showToast('Memutar video selanjutnya...');
    playNextInQueue();
  }
};

function toggleAutoplay(val) {
  autoplayEnabled = val;
  localStorage.setItem('fxtube_autoplay', val);
  showToast(val ? 'Autoplay aktif' : 'Autoplay nonaktif');
}

function renderQueueList() {
  const qList = document.getElementById('queueList');
  if (!qList) return;
  qList.innerHTML = '';
  const subtitle = document.getElementById('queueSubtitle');
  if (subtitle && currentQueue.length > 0) {
    const isRealPlaylist = currentVideo && currentVideo.url && currentVideo.url.includes('list=') && !currentVideo.url.includes('list=RD');
    subtitle.innerText = `${isRealPlaylist ? 'Playlist' : 'YouTube Mix'}: Video ${currentQueueIndex + 1} dari ${currentQueue.length}`;
  }

  currentQueue.forEach((item, idx) => {
    const div = document.createElement('div');
    div.className = 'queue-item' + (idx === currentQueueIndex ? ' active' : '');
    div.onclick = () => {
      if (isMpvActive || playerMode === 'mpv') {
        playDirectMPV(idx, false);
      } else {
        playDirectEmbedded(idx, false);
      }
    };

    let thumb = 'https://i.ytimg.com/img/no_thumbnail.jpg';
    if (item.thumbnails && item.thumbnails.length > 0) {
      thumb = item.thumbnails[0].url || thumb;
    }

    div.innerHTML = `
      <div class="queue-thumb">
        <img src="${thumb}" loading="lazy" decoding="async">
      </div>
      <div class="queue-info">
        <div class="queue-item-title" title="${escapeHtml(item.title)}">${escapeHtml(item.title)}</div>
        <div class="queue-item-channel">${escapeHtml(item.channel || item.uploader || '')}</div>
        ${idx === currentQueueIndex ? '<div class="queue-now-badge">SEDANG DIPUTAR</div>' : ''}
      </div>
    `;
    qList.appendChild(div);
  });

  const activeEl = qList.querySelector('.queue-item.active');
  if (activeEl) {
    activeEl.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
  }
}

// --- POP OUT MPV ---
async function popOutToMPV() {
  if (!currentVideo) return;
  const idx = currentQueueIndex >= 0 ? currentQueueIndex : 0;
  hidePlayPrompt();
  if (player) {
    player.pause();
    player.removeAttribute('src');
    player.load();
  }
  switchToBrowse();
  playDirectMPV(idx, false);
}

function playPastedUrl() {
  const url = document.getElementById('urlInput').value.trim();
  if (!url) return;
  const v = { url, title: 'Video Kustom', channel: 'Direct Link', duration: 0 };
  currentBrowseVideos = [v];
  if (playerMode === 'mpv') {
    playDirectMPV(0, true);
  } else {
    playDirectEmbedded(0, true);
  }
  document.getElementById('urlInput').value = '';
}

// --- SETTINGS MODAL ---
function openSettingsModal() {
  const pMpv = document.getElementById('playMPV');
  const pWeb = document.getElementById('playWeb');
  if (pMpv && pWeb) {
    if (playerMode === 'web') pWeb.checked = true;
    else pMpv.checked = true;
  }

  const qualityRadios = ['144', '240', '360', '480', '720', '1080'];
  qualityRadios.forEach(q => {
    const el = document.getElementById('set' + q);
    if (el) el.checked = (selectedQuality === q);
  });

  const aGuest = document.getElementById('accGuest');
  const aCustom = document.getElementById('accCustom');
  const aFirefox = document.getElementById('accFirefox');
  if (aGuest && aCustom && aFirefox) {
    if (accountMode === 'guest') aGuest.checked = true;
    else if (accountMode === 'custom') aCustom.checked = true;
    else aFirefox.checked = true;
  }

  const modeBadge = document.getElementById('modeStatusBadge');
  const lanSec = document.getElementById('lanAccessSection');
  if (isStandaloneMode) {
    if (modeBadge) modeBadge.innerText = '🟢 Mode Mandiri (Standalone di HP)';
    if (lanSec) lanSec.style.display = 'none';
  } else {
    if (modeBadge) modeBadge.innerText = '💻 Mode Server Terhubung';
    if (lanSec) lanSec.style.display = 'block';
  }

  // Fetch local LAN network address for mobile access
  if (!isStandaloneMode) {
    fetch('/api/network/info')
      .then(r => r.json())
      .then(data => {
        const code = document.getElementById('mobileLanUrl');
        if (code && data && data.mobile_url) {
          code.innerText = data.mobile_url;
        }
      })
      .catch(() => {});
  }

  const modal = document.getElementById('settingsModal');
  if (modal) modal.style.display = 'flex';
}

function toggleServerConfig() {
  if (window.AndroidApp && typeof window.AndroidApp.openServerConfig === 'function') {
    window.AndroidApp.openServerConfig();
  } else {
    const cur = localStorage.getItem('fxtube_server_url') || '';
    const res = prompt('Masukkan URL Server FX Tube Laptop (misal: http://192.168.1.5:18201):\nKosongkan untuk kembali ke Mode Mandiri:', cur);
    if (res !== null) {
      if (res.trim()) {
        localStorage.setItem('fxtube_server_url', res.trim());
        localStorage.removeItem('fxtube_standalone');
        window.location.href = res.trim();
      } else {
        localStorage.setItem('fxtube_standalone', '1');
        localStorage.removeItem('fxtube_server_url');
        location.reload();
      }
    }
  }
}

function copyMobileUrl() {
  const code = document.getElementById('mobileLanUrl');
  const url = code ? code.innerText : '';
  if (url && navigator.clipboard) {
    navigator.clipboard.writeText(url).then(() => {
      showToast('Link HP disalin: ' + url, '#10b981');
    }).catch(() => {
      showToast('Buka di browser HP: ' + url, '#2563eb');
    });
  } else if (url) {
    showToast('Buka di browser HP: ' + url, '#2563eb');
  }
}

function closeSettingsModal() {
  const modal = document.getElementById('settingsModal');
  if (modal) modal.style.display = 'none';
}

function updateSidebarBadges() {
  const accEl = document.getElementById('accountModeBadge');
  if (accEl) {
    accEl.innerText = accountMode === 'firefox' ? 'Firefox Sync' : (accountMode === 'custom' ? 'Cookies.txt' : 'Guest Mode');
  }
  const pEl = document.getElementById('playerModeBadge');
  if (pEl) {
    pEl.innerText = playerMode === 'mpv' ? 'MPV Native' : 'Embedded Web';
  }
  const qEl = document.getElementById('specOptText');
  if (qEl) {
    qEl.innerText = `${selectedQuality}p H.264`;
  }
}

function saveSettings() {
  const pMode = document.querySelector('input[name="play_mode"]:checked')?.value || playerMode || 'web';
  const q = document.querySelector('input[name="default_quality"]:checked')?.value || selectedQuality || '480';
  const acc = document.querySelector('input[name="acc_mode"]:checked')?.value || accountMode || 'firefox';
  const auto = document.getElementById('setAutoplay')?.checked ?? autoplayEnabled;

  playerMode = pMode;
  selectedQuality = q;
  autoplayEnabled = auto;
  accountMode = acc;

  localStorage.setItem('fxtube_player_mode', pMode);
  localStorage.setItem('fxtube_quality', q);
  localStorage.setItem('fxtube_autoplay', auto);
  localStorage.setItem('fxtube_account_mode', acc);

  closeSettingsModal();
  showToast(`Pengaturan disimpan: Mode ${pMode.toUpperCase()}`);

  const chkAuto = document.getElementById('chkAutoplay');
  if (chkAuto) chkAuto.checked = auto;
  renderQualityButtons(currentAvailableQualities);
  updateSidebarBadges();

  fetchWithTimeout('/api/account-mode', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({ mode: acc }),
    timeout: 3000
  }).catch(() => {});

  refreshCurrentTab();
}

function setupInfiniteScroll() {
  const sentinel = document.getElementById('infiniteSentinel');
  if (!sentinel) return;
  const observer = new IntersectionObserver((entries) => {
    if (entries[0].isIntersecting && !isFetchingMore && hasMoreVideos && document.getElementById('browseView')?.style.display !== 'none') {
      loadMoreVideos();
    }
  }, { rootMargin: '300px' });
  observer.observe(sentinel);
}

window.onload = async () => {
  const chkAuto = document.getElementById('chkAutoplay');
  if (chkAuto) chkAuto.checked = autoplayEnabled;
  renderQualityButtons(currentAvailableQualities);
  updateSidebarBadges();
  setupInfiniteScroll();

  const modal = document.getElementById('settingsModal');
  if (modal) {
    modal.addEventListener('click', (e) => {
      if (e.target === modal) closeSettingsModal();
    });
  }

  // Keyboard Shortcuts
  document.addEventListener('keydown', (e) => {
    const isInput = ['INPUT', 'TEXTAREA'].includes(document.activeElement?.tagName);
    if (isInput) {
      if (e.key === 'Escape') {
        document.activeElement.blur();
        closeSettingsModal();
      }
      return;
    }

    if (e.key === 'Escape') {
      const modalEl = document.getElementById('settingsModal');
      if (modalEl && modalEl.style.display === 'flex') {
        closeSettingsModal();
      } else {
        switchToBrowse();
      }
    } else if (e.key === '/') {
      e.preventDefault();
      document.getElementById('searchInput')?.focus();
    } else if (e.key === ' ') {
      e.preventDefault();
      if (isMpvActive) {
        toggleMPVPause();
      } else if (player && document.getElementById('watchView')?.style.display === 'block') {
        if (player.paused) player.play().catch(() => {}); else player.pause();
      }
    } else if (e.key === 'ArrowRight') {
      if (isMpvActive) seekMPV(10);
      else if (player && document.getElementById('watchView')?.style.display === 'block') {
        seekWeb(10);
      }
    } else if (e.key === 'ArrowLeft') {
      if (isMpvActive) seekMPV(-10);
      else if (player && document.getElementById('watchView')?.style.display === 'block') {
        seekWeb(-10);
      }
    }
  });

  try {
    await fetchWithTimeout('/api/account-mode', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({ mode: accountMode }),
      timeout: 3000
    });
  } catch (e) {}

  loadTab('home');
};
