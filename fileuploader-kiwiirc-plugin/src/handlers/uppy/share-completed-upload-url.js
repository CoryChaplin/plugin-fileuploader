import { friendlyUrl } from '../../utils/friendly-url';
import { decodeMetadata } from '../../utils/decode-metadata';

function headWithRetry(url, retries = 3, delay = 1000) {
    return fetch(url, { method: 'HEAD' }).then((resp) => {
        if ((resp.status === 200 || resp.status === 412) || retries <= 0) {
            return resp;
        }
        return new Promise((resolve) => setTimeout(resolve, delay))
            .then(() => headWithRetry(url, retries - 1, delay));
    }).catch((err) => {
        if (retries <= 0) throw err;
        return new Promise((resolve) => setTimeout(resolve, delay))
            .then(() => headWithRetry(url, retries - 1, delay));
    });
}

export function shareCompletedUploadUrl(kiwiApi) {
    return function handleUploadSuccess(file, response) {
        const url = friendlyUrl(file, response);
        headWithRetry(url).then((headResp) => {
            if (headResp.status !== 200 && headResp.status !== 412) {
                // old server instance responds with 412 Precondition Failed
                return;
            }
            sendUploadEvent(kiwiApi, url, file, headResp);
        }).catch(() => {
            // Share anyway after all retries are exhausted
            sendUploadEvent(kiwiApi, url, file);
        });
    };
}

function sendUploadEvent(kiwiApi, url, file, headResp) {
    let rawMetadata;
    if (headResp) {
        rawMetadata = headResp.headers.get('upload-metadata');
    }

    let metadata = {};
    if (rawMetadata) {
        metadata = decodeMetadata(rawMetadata);
    }

    // emit a global kiwi event
    kiwiApi.emit('fileuploader.uploaded', { url, file, metadata });

    const buffer = file.kiwiFileUploaderTargetBuffer;
    const tagData = {
        size: file.size,
        type: file.type,
    };
    if (metadata.expires) {
        tagData.expires = parseInt(metadata.expires, 10);
    }

    const msgTemplate = kiwiApi.state.getSetting('settings.fileuploader.uploadMessage');
    const message = msgTemplate.replace('%URL%', url);

    buffer.say(message, {
        tags: { '+kiwiirc.com/fileuploader': JSON.stringify(tagData) },
    });
}
