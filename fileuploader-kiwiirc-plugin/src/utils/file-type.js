export const IMAGE_EXTS = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp', 'ico', 'avif'];
export const VIDEO_EXTS = ['mp4', 'webm', 'ogv', 'mov', 'mkv'];
export const AUDIO_EXTS = ['mp3', 'wav', 'flac', 'ogg', 'oga', 'aac', 'opus', 'm4a'];
export const TEXT_EXTS = ['txt', 'log', 'md', 'json', 'xml', 'csv', 'ini', 'cfg', 'py', 'js', 'sh', 'html', 'css', 'yml', 'yaml'];

export function getTypeByExt(url) {
    let pathname;
    try { pathname = new URL(url).pathname; } catch (e) { return null; }
    const ext = pathname.split('.').pop().split('?')[0].toLowerCase();
    if (IMAGE_EXTS.includes(ext)) return 'image';
    if (VIDEO_EXTS.includes(ext)) return 'video';
    if (AUDIO_EXTS.includes(ext)) return 'audio';
    if (TEXT_EXTS.includes(ext)) return 'text';
    return null;
}

export function getTypeByMime(ct) {
    ct = (ct || '').split(';')[0].trim().toLowerCase();
    if (ct.startsWith('image/')) return 'image';
    if (ct.startsWith('video/')) return 'video';
    if (ct.startsWith('audio/')) return 'audio';
    if (ct === 'text/plain' || ct === 'application/json' || ct.includes('xml')) return 'text';
    return 'other';
}
