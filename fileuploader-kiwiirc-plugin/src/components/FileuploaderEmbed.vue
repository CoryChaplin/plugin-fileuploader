<template>
    <div class="kiwi-fileuploader-embed">
        <div v-if="loading" class="kiwi-fileuploader-embed-loading">
            {{ t('loading') }}
        </div>
        <template v-else>
            <!-- IMAGE -->
            <div v-if="contentType === 'image'" class="kiwi-fileuploader-embed-card kiwi-fileuploader-embed-image">
                <a :href="url" target="_blank" rel="noopener noreferrer">
                    <img :src="rawUrl" @load="onMediaReady" @error="onMediaError">
                </a>
            </div>

            <!-- VIDEO -->
            <div v-else-if="contentType === 'video'" class="kiwi-fileuploader-embed-card kiwi-fileuploader-embed-video">
                <video controls preload="metadata" @loadedmetadata="onMediaReady" @error="onMediaError">
                    <source :src="rawUrl">
                    {{ t('video_not_supported') }}
                </video>
            </div>

            <!-- AUDIO -->
            <div v-else-if="contentType === 'audio'" class="kiwi-fileuploader-embed-card kiwi-fileuploader-embed-audio">
                <audio controls preload="metadata" @canplay="onMediaReady" @error="onMediaError">
                    <source :src="rawUrl">
                    {{ t('audio_not_supported') }}
                </audio>
            </div>

            <!-- TEXT -->
            <div v-else-if="contentType === 'text'" class="kiwi-fileuploader-embed-card kiwi-fileuploader-embed-text">
                <pre><code>{{ textContent }}</code></pre>
                <a :href="url" target="_blank" rel="noopener noreferrer" class="kiwi-fileuploader-embed-source-link">{{ url }}</a>
            </div>

            <!-- OTHER / FALLBACK -->
            <div v-else class="kiwi-fileuploader-embed-card kiwi-fileuploader-embed-link">
                <i class="fa fa-file-o" aria-hidden="true" />
                <a :href="url" target="_blank" rel="noopener noreferrer" class="kiwi-fileuploader-embed-page-title">
                    {{ pageTitle || url }}
                </a>
            </div>
        </template>
    </div>
</template>

<script>
'kiwi public';

/* global kiwi:true */
import { getTypeByExt, getTypeByMime } from '@/utils/file-type';

export default {
    props: ['url', 'showPin', 'iframeSandboxOptions'],

    data() {
        return {
            loading: true,
            contentType: 'other',
            textContent: '',
            pageTitle: '',
        };
    },

    computed: {
        rawUrl() {
            try {
                const u = new URL(this.url);
                if (u.protocol !== 'https:' && u.protocol !== 'http:') {
                    return '';
                }
                u.searchParams.set('raw', '1');
                return u.toString();
            } catch (e) {
                return '';
            }
        },
    },

    watch: {
        loading(newVal) {
            if (!newVal) {
                this.$nextTick(() => this.applyHeight());
            }
        },
        url() {
            this.detectContent();
        },
    },

    created() {
        this.detectContent();
    },

    methods: {
        t(key) {
            return kiwi.i18n.t(key, { ns: 'plugin-fileuploader' });
        },
        detectContent() {
            this.loading = true;
            this.contentType = 'other';
            this.textContent = '';
            this.pageTitle = '';

            // Fast path: detect by file extension (no network request needed)
            const extType = getTypeByExt(this.url);
            if (extType) {
                this.contentType = extType;
                if (extType === 'text') {
                    this.fetchText();
                } else {
                    this.loading = false;
                }
                return;
            }

            // Slow path: HEAD request to read Content-Type
            fetch(this.rawUrl, { method: 'HEAD' })
                .then((resp) => {
                    const ct = resp.headers.get('content-type');
                    this.contentType = getTypeByMime(ct);
                    if (this.contentType === 'text') {
                        return this.fetchText();
                    }
                    if (this.contentType === 'other') {
                        return this.fetchPageTitle();
                    }
                    this.loading = false;
                })
                .catch(() => {
                    this.contentType = 'other';
                    this.loading = false;
                });
        },

        fetchText() {
            return fetch(this.rawUrl)
                .then((resp) => resp.text())
                .then((txt) => {
                    this.textContent = txt.length > 5000 ? txt.substring(0, 5000) + '\u2026' : txt;
                    this.loading = false;
                })
                .catch(() => {
                    this.textContent = '';
                    this.loading = false;
                });
        },

        fetchPageTitle() {
            return fetch(this.url)
                .then((resp) => resp.text())
                .then((html) => {
                    const doc = new DOMParser().parseFromString(html, 'text/html');
                    this.pageTitle = (doc.title || '').replace(/\s+/g, ' ').trim();
                    this.loading = false;
                })
                .catch(() => {
                    this.pageTitle = '';
                    this.loading = false;
                });
        },

        applyHeight() {
            this.$emit('setHeight', 'auto');
            if (this.showPin) {
                if (this.$el) this.$el.style.maxHeight = '400px';
            } else {
                this.$emit('setMaxHeight', '54%');
            }
        },

        onMediaReady() {
            this.applyHeight();
        },

        onMediaError() {
            if (this.showPin) {
                this.$emit('close');
            }
        },
    },
};
</script>

<style>
.kiwi-fileuploader-embed {
    padding: 4px 0;
}

.kiwi-fileuploader-embed-loading {
    padding: 8px;
    font-size: 0.9em;
    opacity: 0.6;
}

/* Shared card container */
.kiwi-fileuploader-embed-card {
    display: inline-block;
    max-width: 100%;
    border: 1px solid rgba(0, 0, 0, 0.15);
    border-radius: 5px;
    overflow: hidden;
    background: var(--brand-default-bg, #fff);
    vertical-align: top;
}

/* IMAGE */
.kiwi-fileuploader-embed-image a {
    display: block;
    line-height: 0;
}

.kiwi-fileuploader-embed-image img {
    display: block;
    max-width: 600px;
    max-height: 400px;
    object-fit: contain;
}

/* VIDEO */
.kiwi-fileuploader-embed-video video {
    display: block;
    width: 100%;
    max-width: 600px;
    max-height: 340px;
}

/* AUDIO */
.kiwi-fileuploader-embed-audio audio {
    display: block;
    width: 100%;
    max-width: 500px;
    padding: 6px;
    box-sizing: border-box;
}

/* TEXT */
.kiwi-fileuploader-embed-text {
    max-width: 100%;
}

.kiwi-fileuploader-embed-text pre {
    margin: 0;
    padding: 10px 14px;
    overflow: auto;
    max-height: 300px;
    font-family: monospace;
    font-size: 0.82em;
    line-height: 1.45;
    background: var(--brand-input-bg, #f6f8fa);
    white-space: pre-wrap;
    word-break: break-all;
    border-bottom: 1px solid rgba(0, 0, 0, 0.1);
}

.kiwi-fileuploader-embed-text code {
    display: block;
}

.kiwi-fileuploader-embed-source-link {
    display: block;
    padding: 5px 10px;
    font-size: 0.8em;
    color: var(--brand-primary, #42b983);
    text-decoration: none;
    word-break: break-all;
}

/* LINK / OTHER */
.kiwi-fileuploader-embed-link {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 10px 14px;
    max-width: 500px;
}

.kiwi-fileuploader-embed-link i {
    font-size: 1.1em;
    opacity: 0.55;
    flex-shrink: 0;
}

.kiwi-fileuploader-embed-page-title {
    font-size: 0.95em;
    font-weight: 500;
    color: inherit;
    text-decoration: none;
    word-break: break-word;
}

.kiwi-fileuploader-embed-page-title:hover {
    text-decoration: underline;
}

/* In the main (pinned) media viewer, stretch cards full width */
.kiwi-main-mediaviewer .kiwi-fileuploader-embed-card {
    display: block;
    max-width: none;
    width: 100%;
    border-radius: 0;
    border: none;
}

.kiwi-main-mediaviewer .kiwi-fileuploader-embed-image img {
    max-height: none;
    max-width: 100%;
}

.kiwi-main-mediaviewer .kiwi-fileuploader-embed-video video {
    max-width: 100%;
    max-height: none;
}

.kiwi-main-mediaviewer .kiwi-fileuploader-embed-audio audio {
    max-width: 100%;
}

.kiwi-main-mediaviewer .kiwi-fileuploader-embed-text {
    width: 100%;
    max-width: none;
}
</style>
