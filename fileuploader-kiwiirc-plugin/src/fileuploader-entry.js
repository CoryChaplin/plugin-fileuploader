// polyfill globals for uppy on IE11
import 'core-js/features/array/iterator';
// import 'core-js/features/promise' // already included by kiwiirc

import '@uppy/core/dist/style.css';
import '@uppy/dashboard/dist/style.css';
import '@uppy/webcam/dist/style.css';
import '@uppy/audio/dist/style.css';
import '@uppy/image-editor/dist/style.css';

import FileuploaderEmbed from './components/FileuploaderEmbed.vue';
import sidebarFileList from './components/SidebarFileList.vue';
import { MiB } from './constants/data-size';
import { showDashboardOnDragEnter } from './handlers/show-dashboard-on-drag-enter';
import { uploadOnPaste } from './handlers/upload-on-paste';
import { closeModalWhenUploadsCompleted } from './handlers/uppy/close-modal-when-uploads-completed';
import { shareCompletedUploadUrl } from './handlers/uppy/share-completed-upload-url';
// import { trackFileUploadTarget } from './handlers/uppy/track-file-upload-target';
import instantiateUppy from './instantiate-uppy';
import instantiateUppyLocales from './instantiate-uppy-locales';
import { createPromptUpload } from './prompt-upload';
import { createUploadBlob } from './upload-blob';
import TokenManager from './token-manager';
import { setDefaultSetting } from './utils/set-default-setting';

let scriptPath;

(function() {
    const scriptElements = document.getElementsByTagName('script');
    const thisScriptSrc = scriptElements[scriptElements.length - 1].src;
    scriptPath = thisScriptSrc.substring(0, thisScriptSrc.lastIndexOf('/') + 1);
})();

/* global kiwi:true */
kiwi.plugin('fileuploader', function(kiwiApi, log) {
    // register plugin translations
    kiwiApi.addTranslations('plugin-fileuploader', {
        'en-us': {
            loading: 'Loading\u2026',
            video_not_supported: 'Your browser does not support video playback.',
            audio_not_supported: 'Your browser does not support audio playback.',
            no_files_uploaded: 'No files have recently been uploaded...',
            preview_file: 'Preview File',
            download_file: 'Download File',
            shared_files: 'Shared Files',
            paste_upload_prompt: 'You pasted a lot of text.\nWould you like to upload as a file instead?',
            invalid_upload_target: 'Files can only be shared in channels or queries.',
            upload_message: 'Uploaded file: %URL%',
        },
        'fr-fr': {
            loading: 'Chargement\u2026',
            video_not_supported: 'Votre navigateur ne supporte pas la lecture vid\u00e9o.',
            audio_not_supported: 'Votre navigateur ne supporte pas la lecture audio.',
            no_files_uploaded: 'Aucun fichier n\u2019a \u00e9t\u00e9 partag\u00e9 r\u00e9cemment\u2026',
            preview_file: 'Aper\u00e7u du fichier',
            download_file: 'T\u00e9l\u00e9charger le fichier',
            shared_files: 'Fichiers partag\u00e9s',
            paste_upload_prompt: 'Vous avez coll\u00e9 beaucoup de texte.\nVoulez-vous l\u2019envoyer en tant que fichier ?',
            invalid_upload_target: 'Les fichiers ne peuvent \u00eatre partag\u00e9s que dans des salons ou des conversations priv\u00e9es.',
            upload_message: 'Nouveau fichier : %URL%',
        },
    });

    // default settings
    setDefaultSetting(kiwiApi, 'fileuploader.allowedFileTypes', null);
    setDefaultSetting(kiwiApi, 'fileuploader.maxFileSize', 10 * MiB);
    setDefaultSetting(kiwiApi, 'fileuploader.server', '/files/');
    setDefaultSetting(kiwiApi, 'fileuploader.textPastePromptMinimumLines', 5);
    setDefaultSetting(kiwiApi, 'fileuploader.textPasteNeverPrompt', false);
    setDefaultSetting(kiwiApi, 'fileuploader.bufferInfoUploads', true);
    setDefaultSetting(kiwiApi, 'fileuploader.localePath', '');
    setDefaultSetting(kiwiApi, 'fileuploader.uploadMessage', kiwiApi.i18n.t('upload_message', { ns: 'plugin-fileuploader' }));

    // add button to input bar
    const uploadFileButton = document.createElement('i');
    uploadFileButton.className = 'upload-file-button fa fa-upload';
    kiwiApi.addUi('input', uploadFileButton);

    // add sidebar panel
    if (kiwiApi.state.setting('fileuploader.bufferInfoUploads')) {
        const sidebarComponent = new kiwiApi.Vue(sidebarFileList);
        sidebarComponent.$mount();
        kiwiApi.addUi('about_buffer', sidebarComponent.$el, { title: kiwiApi.i18n.t('shared_files', { ns: 'plugin-fileuploader' }) });
    }

    // set up main uppy object
    const tokenManager = new TokenManager(kiwiApi);
    const { uppy, dashboard } = instantiateUppy({
        kiwiApi,
        tokenManager,
        uploadFileButton,
    });

    instantiateUppyLocales(kiwiApi, uppy, scriptPath);

    const promptUpload = createPromptUpload({ kiwiApi, tokenManager });
    const uploadBlob = createUploadBlob(kiwiApi);
    // expose plugin api
    kiwiApi.fileuploader = { uppy, dashboard, promptUpload, uploadBlob };

    // show uppy modal whenever a file is dragged over the page
    window.addEventListener('dragenter', showDashboardOnDragEnter(kiwiApi, dashboard));

    // show uppy modal when files are pasted
    kiwiApi.on('buffer.paste', uploadOnPaste(kiwiApi, uppy, dashboard));

    // send message with link to buffer when upload finishes
    uppy.on('upload-success', shareCompletedUploadUrl(kiwiApi));

    // hide dashboard after last upload finishes
    uppy.on('complete', closeModalWhenUploadsCompleted(uppy, dashboard));

    // --- URL Embed replacement for fileuploader links ---
    let serverBase = kiwiApi.state.getSetting('settings.fileuploader.server') || '';
    // Resolve relative paths (e.g. '/files/') to absolute URLs
    try { serverBase = new URL(serverBase, window.location.origin).toString(); } catch (e) { /* noop */ }

    if (serverBase) {
        const origUrlEmbed = kiwiApi.require('components/UrlEmbed');
        const OriginalUrlEmbed = Object.assign({}, origUrlEmbed);

        const FileuploaderEmbedWrapper = {
            functional: true,
            props: ['url', 'showPin', 'iframeSandboxOptions'],
            render(h, ctx) {
                const url = ctx.props.url || '';
                const isFileuploader = url.startsWith(serverBase);
                const component = isFileuploader ? FileuploaderEmbed : OriginalUrlEmbed;
                return h(component, {
                    props: ctx.props,
                    on: ctx.listeners,
                });
            },
        };

        kiwiApi.replaceModule('components/UrlEmbed', FileuploaderEmbedWrapper);
    }
});
