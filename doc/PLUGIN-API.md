# Fileuploader plugin API

Other Kiwi IRC plugins can interact with fileuploader through `kiwiApi.fileuploader`, which is populated once the `fileuploader` plugin has loaded.

## Availability

`kiwiApi.fileuploader` is only present after the fileuploader plugin has initialised. If your plugin may load before fileuploader, guard against this:

```js
kiwi.plugin('my-plugin', function(kiwiApi) {
    kiwiApi.on('plugin.loaded', ({ name }) => {
        if (name !== 'fileuploader') return;
        // safe to use kiwiApi.fileuploader here
    });
});
```

---

## Methods

### `kiwiApi.fileuploader.uploadBlob(content, options)` → `Promise<{ url }>`

Programmatically upload a string or binary blob without any UI. The promise resolves with the public file URL once the upload is complete, or rejects on network/server error.

**Parameters**

| Name | Type | Description |
|---|---|---|
| `content` | `string \| Blob \| File` | The content to upload. Strings are wrapped in a UTF-8 `text/plain` Blob unless `mimeType` overrides the type. |
| `options.filename` | `string` | File name to use for the upload. Shown in the HTML preview page. Default: `'upload'`. |
| `options.mimeType` | `string` | MIME type. Determines which preview the server renders (image, video, text, …). Defaults to the Blob's own type, or `'text/plain'` for strings, or `'application/octet-stream'` as a last resort. |
| `options.category` | `string` | Upload category key, as configured in `[Categories.<name>]` on the server. Drives retention duration and storage subtree. Omit for default retention. |

**Returns**

```js
{ url: 'https://files.example.com/files/<id>/filename.txt' }
```

**Example — upload a plain-text abuse report log**

```js
const log = buildConversationLog(); // returns a string

kiwiApi.fileuploader.uploadBlob(log, {
    filename: 'report.log',
    mimeType: 'text/plain',
    category: 'abuse-report',
}).then(({ url }) => {
    // send url to the moderation system
}).catch((err) => {
    console.error('Upload failed:', err);
});
```

**Notes**

- The upload uses a fresh headless Uppy/TUS instance; it does not interfere with any files the user has queued in the dashboard.
- JWT authentication is not attached to programmatic uploads in v1. If the server has `RequireJwtAccount = true`, these uploads will be rejected — use a category with its own `MaxAge` instead of relying on `IdentifiedMaxAge`.
- The server must have the category name registered in `[Categories.<name>]` or the upload will fail with HTTP 400.

---

### `kiwiApi.fileuploader.promptUpload(uppyOptions?)` → `Promise<UppyCompleteEvent>`

Open the fileuploader dashboard modal and let the user pick and upload files. Resolves with the Uppy `complete` event when all uploads finish, or rejects with `'Upload dialog closed'` if the user dismisses the modal.

```js
kiwiApi.fileuploader.promptUpload().then((result) => {
    console.log('uploads complete', result.successful);
});
```

`uppyOptions` is forwarded to the `instantiateUppy` helper and can override `dashboardOptions`, `tusOptions`, or `uppyOptions`. In most cases you do not need to pass anything.

---

## Events

### `fileuploader.uploaded`

Emitted on `kiwiApi` whenever any upload from the fileuploader dashboard completes successfully (one event per file).

**Payload**

```js
{
    url:      string,   // public URL of the uploaded file
    file:     UppyFile, // Uppy file object (name, size, type, meta, …)
    metadata: object,   // decoded Upload-Metadata response headers from the server
                        // e.g. { filename, filetype, expires }
}
```

**Example**

```js
kiwiApi.on('fileuploader.uploaded', ({ url, file, metadata }) => {
    console.log(`${file.meta.name} uploaded to ${url}, expires ${metadata.expires}`);
});
```

This event is **not** fired for programmatic uploads made via `uploadBlob` — use the returned Promise instead.

---

## Direct Uppy access

### `kiwiApi.fileuploader.uppy`

The Uppy instance that backs the dashboard. Use this to listen for low-level Uppy events or inspect queued files. Avoid calling `uppy.upload()` directly — prefer `promptUpload()`.

### `kiwiApi.fileuploader.dashboard`

The Uppy Dashboard plugin instance. You can open or close it programmatically:

```js
kiwiApi.fileuploader.dashboard.openModal();
kiwiApi.fileuploader.dashboard.closeModal();
```

---

## Server-side category configuration

To use a named category in `uploadBlob`, add a section to the server config:

```toml
[Categories.abuse-report]
MaxAge           = "8760h"  # 1 year (applies to anonymous uploads)
IdentifiedMaxAge = "8760h"  # 1 year (applies to JWT-authenticated uploads)
StoragePrefix    = "reports"
# Files land at <Storage.Path>/reports/complete/... instead of
# <Storage.Path>/complete/... — safe from cronjobs that prune the default tree.
```

Uploads with a `category` value not present in the config are rejected with HTTP 400. The empty category (no `category` metadata field) always uses the global `[Expiration]` defaults.

> **Trust note:** the `category` field is client-supplied and not cryptographically verified in v1. Any Kiwi user who can reach the upload endpoint can request any configured category. Size up your storage quotas accordingly.
