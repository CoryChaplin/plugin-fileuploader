import Uppy from '@uppy/core';
import Tus from '@uppy/tus';

import { KiB } from './constants/data-size';
import { friendlyUrl } from './utils/friendly-url';

/**
 * Creates an uploadBlob function bound to the given kiwiApi.
 *
 * uploadBlob(content, options) → Promise<{ url }>
 *
 * content   - string | Blob | File to upload
 * options:
 *   filename  - string, name to use for the upload (default: 'upload')
 *   mimeType  - string, MIME type (default: 'text/plain' for strings, Blob.type otherwise)
 *   category  - string, upload category key from server config (e.g. 'abuse-report')
 *
 * The returned URL is the full file URL with the filename appended, matching
 * the format used by regular fileuploader uploads.
 *
 * NOTE: JWT authentication is not acquired for programmatic uploads in v1.
 * Category-based retention does not require a JWT.
 */
export function createUploadBlob(kiwiApi) {
    return function uploadBlob(content, { filename, mimeType, category } = {}) {
        return new Promise((resolve, reject) => {
            let blob;
            if (typeof content === 'string') {
                blob = new Blob([content], { type: mimeType || 'text/plain' });
            } else if (content instanceof Blob) {
                blob = mimeType ? new Blob([content], { type: mimeType }) : content;
            } else {
                reject(new TypeError('uploadBlob: content must be a string or Blob'));
                return;
            }

            const effectiveMimeType = blob.type || 'application/octet-stream';
            const effectiveFilename = filename || 'upload';

            const uppy = new Uppy({ autoProceed: true })
                .use(Tus, {
                    endpoint: kiwiApi.state.setting('fileuploader.server'),
                    chunkSize: 512 * KiB,
                });

            uppy.on('upload-success', (file, response) => {
                uppy.close();
                resolve({ url: friendlyUrl(file, response) });
            });

            uppy.on('upload-error', (file, error) => {
                uppy.close();
                reject(error);
            });

            uppy.addFile({
                name: effectiveFilename,
                type: effectiveMimeType,
                data: blob,
                meta: {
                    filename: effectiveFilename,
                    filetype: effectiveMimeType,
                    ...(category ? { category } : {}),
                },
            });
        });
    };
}
